import json
import re
from datetime import datetime, timezone

import psycopg2
import scrapy
from scrapy.spidermiddlewares.httperror import HttpError

from items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price

PAGE_SIZE = 24
BASE_URL = "https://fptshop.com.vn"
API_URL = "https://papi.fptshop.com.vn/gw/v1/public/fulltext-search-service/category"
API_HEADERS = {
    "Accept": "application/json",
    "Content-Type": "application/json",
    "order-channel": "1",
    "Origin": BASE_URL,
    "Referer": f"{BASE_URL}/",
}

EXCLUDE_PATHS = {
    "gio-hang",
    "tim-kiem",
    "cua-hang",
    "tos",
    "sim-fpt",
    "may-doi-tra",
    "sim-so-dep",
    "collection",
    "ho-tro",
    "dich-vu",
    "cdn-cgi",
}


def strip_html(text):
    return re.sub(r"<[^>]+>", " ", text).strip() if text else text


class FptSpider(scrapy.Spider):
    name = "fptshop"
    allowed_domains = ["fptshop.com.vn", "papi.fptshop.com.vn"]

    custom_settings = {
        "DOWNLOAD_DELAY": 1.5,
        "RANDOMIZE_DOWNLOAD_DELAY": True,
        "DUPEFILTER_CLASS": "scrapy.dupefilters.RFPDupeFilter",
        "DEFAULT_REQUEST_HEADERS": {
            "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
            "Accept-Language": "vi-VN,vi;q=0.9,en-US;q=0.8,en;q=0.7",
            "User-Agent": (
                "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                "AppleWebKit/537.36 (KHTML, like Gecko) "
                "Chrome/122.0.0.0 Safari/537.36"
            ),
        },
    }

    async def start(self):
        if hasattr(self, "start_url"):
            path = self.start_url.rstrip("/").replace(BASE_URL, "")
            parts = [p for p in path.split("/") if p]
            if len(parts) >= 2:
                yield scrapy.Request(
                    self.start_url, callback=self.parse_product, errback=self.handle_error
                )
            else:
                yield scrapy.Request(
                    API_URL,
                    method="POST",
                    body=json.dumps(
                        {
                            "skipCount": 0,
                            "maxResultCount": PAGE_SIZE,
                            "sortMethod": "noi-bat",
                            "slug": parts[0],
                            "categoryType": "category",
                        }
                    ),
                    headers=API_HEADERS,
                    callback=self.parse_listing,
                    errback=self.handle_error,
                    meta={"slug": parts[0], "skip": 0},
                )
            return

        yield scrapy.Request(
            BASE_URL,
            callback=self.parse_categories,
            errback=self.handle_error,
        )

        db_url = self.settings.get("DATABASE_URL")
        if db_url:
            try:
                conn = psycopg2.connect(db_url)
                cur = conn.cursor()
                cur.execute(
                    "SELECT url FROM products WHERE source = 'fptshop' AND (price IS NULL OR in_stock = false)"
                )
                for (url,) in cur.fetchall():
                    yield scrapy.Request(
                        url,
                        callback=self.parse_product,
                        errback=self.handle_error,
                    )
                cur.close()
                conn.close()
            except Exception as e:
                self.logger.error("Failed to fetch stale URLs from DB: %s", e)

    def parse_categories(self, response):
        seen = set()
        for href in response.css("a::attr(href)").getall():
            if not href or not href.startswith("/"):
                continue
            path = href.strip("/")
            if not path or "?" in path or "#" in path:
                continue
            if path.count("/") > 1:
                continue
            top = path.split("/")[0]
            if top in EXCLUDE_PATHS:
                continue
            if path in seen:
                continue
            seen.add(path)
            yield scrapy.Request(
                API_URL,
                method="POST",
                body=json.dumps(
                    {
                        "skipCount": 0,
                        "maxResultCount": 1,
                        "sortMethod": "noi-bat",
                        "slug": path,
                        "categoryType": "category",
                    }
                ),
                headers=API_HEADERS,
                callback=self.check_category,
                errback=self.handle_error,
                meta={"slug": path, "dont_redirect": True},
            )

    def check_category(self, response):
        data = response.json()
        if not data.get("validSlug") or not data.get("totalCount", 0):
            return
        slug = response.meta["slug"]
        self.logger.debug("Valid category found: %s (%d items)", slug, data["totalCount"])
        yield scrapy.Request(
            API_URL,
            method="POST",
            body=json.dumps(
                {
                    "skipCount": 0,
                    "maxResultCount": PAGE_SIZE,
                    "sortMethod": "noi-bat",
                    "slug": slug,
                    "categoryType": "category",
                }
            ),
            headers=API_HEADERS,
            callback=self.parse_listing,
            errback=self.handle_error,
            meta={"slug": slug, "skip": 0, "dont_redirect": True},
        )

    def parse_listing(self, response):
        data = response.json()
        total = data.get("totalCount", 0)
        items = data.get("items", [])
        slug = response.meta["slug"]
        skip = response.meta["skip"]

        for item in items:
            product_slug = item.get("slug", "")
            if not product_slug:
                continue
            yield scrapy.Request(
                f"{BASE_URL}/{product_slug}",
                callback=self.parse_product,
                errback=self.handle_error,
                meta={"listing": item},
            )

        MAX_SKIP = 1000 - PAGE_SIZE
        next_skip = skip + PAGE_SIZE
        if next_skip < total and next_skip <= MAX_SKIP:
            yield scrapy.Request(
                API_URL,
                method="POST",
                body=json.dumps(
                    {
                        "skipCount": next_skip,
                        "maxResultCount": PAGE_SIZE,
                        "sortMethod": "noi-bat",
                        "slug": slug,
                        "categoryType": "category",
                    }
                ),
                headers=API_HEADERS,
                callback=self.parse_listing,
                errback=self.handle_error,
                meta={"slug": slug, "skip": next_skip, "dont_redirect": True},
            )

    def parse_product(self, response):
        listing = response.meta.get("listing", {})
        if not listing and response.css("h1.product-name, #detail-product-script").get() is None:
            return

        product_data = {}
        raw = response.css("#detail-product-script::text").get()
        if raw:
            try:
                product_data = json.loads(raw)
            except (json.JSONDecodeError, AttributeError) as e:
                self.logger.warning("Failed to parse product JSON at %s: %s", response.url, e)

        offers = product_data.get("offers", {})
        agg = product_data.get("aggregateRating", {})

        # --- name / sku / brand ---
        name = clean_text(
            product_data.get("name") or listing.get("name") or response.css("h1::text").get()
        )
        sku = product_data.get("sku") or listing.get("code")
        brand = clean_text(
            product_data.get("brand", {}).get("name") or (listing.get("brand") or {}).get("name")
        )

        # --- category / subcategory ---
        category = subcategory = None
        raw_bc = response.css("#breadcrumb-structured-data::text").get()
        if raw_bc:
            try:
                bc = json.loads(raw_bc)
                elements = bc.get("itemListElement", [])
                if len(elements) >= 3:
                    category = next(
                        (x.get("name") for x in elements if x.get("position") == 2), None
                    )
                    subcategory = next(
                        (x.get("name") for x in elements if x.get("position") == 3), None
                    )
                elif len(elements) == 2:
                    category = next(
                        (x.get("name") for x in elements if x.get("position") == 2), None
                    )
            except (json.JSONDecodeError, AttributeError) as e:
                self.logger.warning("Failed to parse breadcrumb at %s: %s", response.url, e)

        if not category:
            url_parts = response.url.split("/")
            category = url_parts[3].replace("-", " ").title() if len(url_parts) > 3 else None

        # --- pricing ---
        if listing:
            # API data is authoritative — use it directly
            price = listing.get("currentPrice") or None
            original_price = listing.get("originalPrice") or None
            discount_percent = listing.get("discountPercentage") or None
        else:
            # fallback to HTML for start_url / DB refresh cases
            price = parse_price(str(offers.get("price", ""))) or None
            original_price = (
                parse_price(response.css("span.line-through::text").get() or "") or None
            )
            discount_percent = parse_discount(
                "".join(response.css("span.text-red-red-7::text").getall()[:2])
            )

        if not price:
            price = None
            original_price = None
            discount_percent = None

        if original_price and price and original_price == price:
            original_price = None

        # reverse-calculate original_price only for HTML fallback path
        if (
            not listing
            and price
            and discount_percent is not None
            and discount_percent < 100
            and original_price is None
        ):
            original_price = round(price / (1 - discount_percent / 100))

        # --- stock ---
        skus = listing.get("skus", [])
        total_inv = listing.get("totalInventory", 0)
        in_stock = (
            total_inv > 0
            or any(s.get("totalInventory", 0) > 0 for s in skus)
            or offers.get("availability", "").endswith("InStock")
        )
        quantity = total_inv or None

        # --- rating ---
        rating = float(agg["ratingValue"]) if agg.get("ratingValue") else None
        review_count = int(float(agg["reviewCount"])) if agg.get("reviewCount") else None

        # --- description ---
        desc_container = response.xpath(
            "//h2[contains(text(),'Mô tả sản phẩm')]/../../following-sibling::div[1]"
        )
        if desc_container:
            elements = desc_container.xpath(".//p | .//h2 | .//h3 | .//li")
            parts = [
                re.sub(r"\s+", " ", strip_html(el.get())).strip()
                for el in elements
                if strip_html(el.get()).strip() not in ("Thu gọn", "")
            ]
            description = "\n\n".join(parts)
        else:
            raw_desc = product_data.get("description")
            description = clean_text(raw_desc.strip('"')) if raw_desc else None

        # --- images ---
        images = list(
            dict.fromkeys(
                url
                for url in response.css("[class*='swiper'] img::attr(src)").getall()
                if "/unsafe/828x0/" in url
            )
        )
        if not images:
            ld_image = product_data.get("image") or (
                listing.get("image", {}).get("src")
                if isinstance(listing.get("image"), dict)
                else listing.get("image")
            )
            if isinstance(ld_image, list):
                images = ld_image
            elif ld_image:
                images = [ld_image]

        # --- specs ---
        specs = {
            prop["name"]: prop["value"]
            for prop in (product_data.get("additionalProperty") or [])
            if prop.get("name") and str(prop.get("value", "")).strip()
        }

        yield ProductItem(
            url=response.url.split("?")[0],
            source="fptshop",
            crawled_at=datetime.now(timezone.utc),
            currency="VND",
            name=name,
            sku=sku,
            brand=brand,
            category=category,
            subcategory=subcategory,
            description=description,
            price=price,
            original_price=original_price,
            discount_percent=discount_percent,
            in_stock=in_stock,
            quantity=quantity,
            rating=rating,
            review_count=review_count,
            images=images,
            specs=specs,
        )

    def handle_error(self, failure):
        if failure.check(HttpError):
            response = failure.value.response
            self.logger.error(
                "HTTP %s for %s — request body: %s — response: %s",
                response.status,
                failure.request.url,
                failure.request.body,
                response.text[:500],
            )
        else:
            self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

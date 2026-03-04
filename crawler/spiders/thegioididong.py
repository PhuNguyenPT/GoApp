import json
import re
from datetime import datetime, timezone

import psycopg2
import scrapy

from items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price, parse_rating

EXCLUDED_NAV = {
    "/sim-so-dep",
    "/tien-ich",
    "/tien-ich-khac",
    "/tin-tuc",
    "/hoi-dap",
    "/tra-gop",
    "/bao-hanh",
    "/cart",
    "/flashsale",
    "/online-only",
    "/thu-cu-doi-moi",
    "/dong-ho-gia-soc",
    "/chuong-trinh-dac-quyen-tgdd",
    "/chuong-trinh-tra-cham-ict",
    "/he-thong-sieu-thi-the-gioi-di-dong",
    "/thong-tin-khac",
    "/dat-ve-may-bay",
    "/may-doi-tra",
}

FALLBACK_URLS = [
    "/dtdd",
    "/laptop",
    "/may-tinh-bang",
    "/may-tinh-de-ban",
    "/tai-nghe",
    "/phu-kien",
    "/dong-ho-thong-minh-ldp",
    "/dong-ho",
    "/man-hinh-may-tinh",
    "/camera-giam-sat",
    "/loa-laptop",
    "/loa",
    "/chuot-may-tinh",
    "/ban-phim",
    "/sac-dtdd",
    "/adapter-sac",
    "/may-in",
    "/muc-in",
]


def strip_html(text):
    return re.sub(r"<[^>]+>", "", text).strip() if text else text


class ThegioididongSpider(scrapy.Spider):
    name = "thegioididong"
    allowed_domains = ["thegioididong.com"]

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
            yield scrapy.Request(self.start_url, callback=self.parse_product)
            return
        yield scrapy.Request(
            "https://www.thegioididong.com",
            callback=self.parse_categories,
            errback=self.handle_error,
        )

        db_url = self.settings.get("DATABASE_URL")
        if db_url:
            try:
                conn = psycopg2.connect(db_url)
                cur = conn.cursor()
                cur.execute(
                    "SELECT url FROM products WHERE source = 'thegioididong' AND (price IS NULL OR in_stock = false)"
                )
                for (url,) in cur.fetchall():
                    yield scrapy.Request(
                        url,
                        callback=self.parse_product,
                        errback=self.handle_error,
                        dont_filter=True,
                    )
                cur.close()
                conn.close()
            except Exception as e:
                self.logger.error("Failed to fetch stale URLs from DB: %s", e)

    def parse_categories(self, response):
        hrefs = [
            h
            for h in response.css("[class*='main'] a::attr(href)").getall()
            if h
            and h.startswith("/")
            and "?" not in h
            and "#" not in h
            and not h.startswith("javascript")
            and h.split("/")[1] not in {e.lstrip("/") for e in EXCLUDED_NAV}
        ]
        if not hrefs:
            self.logger.warning("Nav selector returned nothing — using fallback URLs")
            hrefs = FALLBACK_URLS
        seen = set()
        for href in hrefs:
            if href in seen:
                continue
            seen.add(href)
            yield response.follow(
                href,
                callback=self.parse,
                errback=self.handle_error,
            )

    def parse(self, response):
        for link in response.css("ul.listproduct li.item a.main-contain::attr(href)").getall():
            yield response.follow(
                link,
                callback=self.parse_product,
                errback=self.handle_error,
            )
        next_page = response.css("a.next::attr(href)").get()
        if next_page:
            yield response.follow(
                next_page,
                callback=self.parse,
                errback=self.handle_error,
            )

    def _parse_gtm_fallback(self, response):
        """Extract product data from GTM objGTM variable when LD+JSON Product is missing."""
        all_scripts = response.css("script:not([src])::text").getall()

        # --- Pattern 1: objGTM (landing pages like Xiaomi 17 Ultra) ---
        gtm_scripts = [s for s in all_scripts if "'name'" in s and "'price'" in s and "'id'" in s]
        if gtm_scripts:
            text = gtm_scripts[0]

            def _get(key):
                m = re.search(rf"'{key}':\s*'([^']*)'", text)
                return m.group(1).strip() if m and m.group(1).strip() else None

            price_raw = _get("price")
            return {
                "name": _get("name"),
                "sku": _get("id"),
                "brand": None,
                "price": float(price_raw) if price_raw else None,
                "in_stock": _get("dimension55") == "Yes",
            }

        # --- Pattern 2: GA4 dataLayer / MAddToCartAll (info/coming-soon pages) ---
        ga4_scripts = [s for s in all_scripts if "item_id:" in s and "item_brand:" in s]
        if ga4_scripts:
            text = ga4_scripts[0]

            def _get_ga4(key):
                m = re.search(rf'{key}:\s*["`\']([^"`\']*)["`\']', text)
                return m.group(1).strip() if m and m.group(1).strip() else None

            price_raw = re.search(r"price:\s*([\d.]+)", text)
            price = float(price_raw.group(1)) if price_raw else None

            return {
                "name": response.css(
                    "h1::text"
                ).get(),  # GA4 uses template literals, h1 is reliable
                "sku": _get_ga4("item_id"),
                "brand": _get_ga4("item_brand"),
                "price": price if price and price > 0 else None,
                "in_stock": False,  # price: 0.0 means not yet available
            }

        return {}

    def _parse_gtm_specs(self, response):
        """Extract specs from the inline-styled table used on landing/pre-order pages."""
        specs = {}
        for table in response.css("table"):
            tds = table.css("td")
            # Pairs: even index = label, odd index = value
            for i in range(0, len(tds) - 1, 2):
                key = clean_text(tds[i].css("::text").get())
                val = strip_html(tds[i + 1].get())
                if key and val and val.strip() != "Đang cập nhật":
                    specs[key] = val
            if specs:
                break  # stop at first table with data
        return specs

    def parse_product(self, response):
        ld_product = {}
        ld_breadcrumbs = []
        for script in response.css("script[type='application/ld+json']::text").getall():
            try:
                data = json.loads(script)
                if data.get("@type") == "Product":
                    ld_product = data
                elif data.get("@type") == "BreadcrumbList":
                    ld_breadcrumbs = data.get("itemListElement", [])
            except (json.JSONDecodeError, AttributeError) as e:
                self.logger.warning("Failed to parse LD+JSON at %s: %s", response.url, e)

        offers = ld_product.get("offers") or {}
        rating = ld_product.get("aggregateRating") or {}

        item = ProductItem()
        item["source"] = "thegioididong"
        item["url"] = ld_product.get("url") or response.url.split("?")[0]
        item["crawled_at"] = datetime.now(timezone.utc).isoformat()
        item["currency"] = offers.get("priceCurrency", "VND")

        # --- Fallback for pages without Product LD+JSON (landing/pre-order pages) ---
        if not ld_product:
            gtm = self._parse_gtm_fallback(response)
            item["name"] = gtm.get("name")
            item["sku"] = gtm.get("sku")
            item["description"] = None
            item["brand"] = gtm.get("brand") or ((gtm.get("name") or "").split()[0] or None)
            item["category"] = (
                ld_breadcrumbs[-1].get("item", {}).get("name") if ld_breadcrumbs else None
            )
            item["price"] = gtm.get("price")
            item["original_price"] = None
            item["discount_percent"] = None
            item["in_stock"] = gtm.get("in_stock", False)
            item["quantity"] = None
            item["rating"] = None
            item["review_count"] = None

            # Product images are JS-rendered and not in static HTML on landing pages.
            # Use og:image (product reveal/press image) as best available fallback.
            src_images = [
                url
                for url in response.css("div.owl-carousel img::attr(src)").getall()
                if "/Products/" in url
            ]
            data_src_images = [
                url
                for url in response.css("div.owl-carousel img::attr(data-src)").getall()
                if "/Products/" in url
            ]
            product_images = list(dict.fromkeys(src_images + data_src_images))
            if not product_images:
                og_image = response.css("meta[property='og:image']::attr(content)").get()
                product_images = [og_image.split("#")[0]] if og_image else []
            item["images"] = product_images

            item["specs"] = self._parse_gtm_specs(response)
            yield item
            return

        # --- Normal path: Product LD+JSON present ---
        item["name"] = ld_product.get("name") or clean_text(response.css("h1::text").get())
        item["sku"] = ld_product.get("sku")

        # Long description from article body, fall back to JSON-LD SEO summary
        desc_html = response.css("div.description div.text-detail").get()
        if desc_html:
            item["description"] = re.sub(r"\s+", " ", strip_html(desc_html)).strip()
        else:
            item["description"] = ld_product.get("description")

        # Brand: {"@type": "Brand", "name": ["iPhone (Apple)"]} — name is a list
        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        brand_match = re.search(r"\((.+?)\)", brand_raw) if brand_raw else None
        item["brand"] = brand_match.group(1) if brand_match else brand_raw

        # Last resort: derive brand from first word of product name
        if not item["brand"] and item["name"]:
            item["brand"] = item["name"].split()[0] if item["name"].split() else None

        # Category from breadcrumb JSON-LD (last item)
        if ld_breadcrumbs:
            item["category"] = ld_breadcrumbs[-1].get("item", {}).get("name")
        else:
            breadcrumbs = response.css("[class*='breadcrumb'] a::text").getall()
            item["category"] = clean_text(breadcrumbs[-1]) if breadcrumbs else None

        item["price"] = offers.get("price") or parse_price(
            response.css("[class*='price-present']::text").get()
        )
        item["original_price"] = parse_price(response.css("p.box-price-old::text").get())

        if item["original_price"] and item["price"] and item["original_price"] > item["price"]:
            diff = item["original_price"] - item["price"]
            item["discount_percent"] = round((diff / item["original_price"]) * 100)
        else:
            item["discount_percent"] = parse_discount(
                response.css("p.box-price-percent::text").get()
            )

        if item["price"] and item["discount_percent"] and not item["original_price"]:
            discount = item["discount_percent"]
            item["original_price"] = (
                round(item["price"] / (1 - discount / 100)) if 0 < discount < 100 else None
            )

        availability = offers.get("availability", "")
        item["in_stock"] = "InStock" in availability or bool(
            response.css("[class*='btn-buynow']").get()
        )
        item["quantity"] = None

        item["rating"] = rating.get("ratingValue") or parse_rating(
            response.css(".point-average-score::text").get()
        )
        item["review_count"] = rating.get("reviewCount")

        ld_image = ld_product.get("image", {})
        ld_image_url = ld_image.get("contentUrl") if isinstance(ld_image, dict) else ld_image
        src_images = [
            url
            for url in response.css("div.owl-carousel img::attr(src)").getall()
            if "/Products/" in url
        ]
        data_src_images = [
            url
            for url in response.css("div.owl-carousel img::attr(data-src)").getall()
            if "/Products/" in url
        ]
        product_images = list(dict.fromkeys(src_images + data_src_images))
        item["images"] = product_images or ([ld_image_url] if ld_image_url else [])

        item["specs"] = {
            prop["name"]: strip_html(prop["value"])
            for prop in (ld_product.get("additionalProperty") or [])
            if prop.get("name")
            and prop.get("value")
            and strip_html(prop["value"]).strip() != "Đang cập nhật"
        }

        yield item

    def handle_error(self, failure):
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

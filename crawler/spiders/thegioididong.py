import json
import re
from datetime import datetime, timezone

import scrapy
from scrapy.spidermiddlewares.httperror import HttpError

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
    "/dtdd-samsung",
    "/dtdd-apple-iphone",
    "/dtdd-xiaomi",
    "/dtdd-oppo",
    "/dtdd-vivo",
    "/dtdd-realme",
    "/dtdd-honor",
    "/dtdd-nokia",
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

PAGE_SIZE = 20
BASE_URL = "https://www.thegioididong.com"


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
            path = self.start_url.rstrip("/").replace(BASE_URL, "")
            parts = [p for p in path.split("/") if p]
            if len(parts) >= 2:
                yield scrapy.Request(
                    self.start_url, callback=self.parse_product, errback=self.handle_error
                )
            else:
                yield scrapy.Request(
                    self.start_url, callback=self.parse_category_page, errback=self.handle_error
                )
            return
        yield scrapy.Request(
            BASE_URL,
            callback=self.parse_categories,
            errback=self.handle_error,
        )

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
                callback=self.parse_category_page,
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

    def parse_category_page(self, response):
        cate_id = next(
            (
                v
                for v in response.css("[data-cate]::attr(data-cate)").getall()
                if v.isdigit() and int(v) > 0
            ),
            None,
        )
        if not cate_id:
            self.logger.debug("No cate_id at %s — falling back to CSS pagination", response.url)
            yield from self.parse(response)
            return

        yield scrapy.Request(
            f"{BASE_URL}/Category/FilterProductBox?c={cate_id}&o=13&pi=1",
            method="POST",
            body="IsParentCate=False&IsShowCompare=True&prevent=true",
            headers={
                "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
                "X-Requested-With": "XMLHttpRequest",
                "Referer": response.url,
            },
            callback=self.parse_listing,
            errback=self.handle_error,
            meta={"cate_id": cate_id, "page": 1, "referer": response.url},
        )

    def parse_listing(self, response):
        data = response.json()
        total = data.get("total", 0)
        html = data.get("listproducts", "")
        sel = scrapy.Selector(text=html)

        for href in sel.css("li.item a.main-contain::attr(href)").getall():
            yield response.follow(
                href.split("?")[0],
                callback=self.parse_product,
                errback=self.handle_error,
            )

        page = response.meta["page"]
        cate_id = response.meta["cate_id"]
        referer = response.meta["referer"]

        if page * PAGE_SIZE < total:
            next_page = page + 1
            yield scrapy.Request(
                f"{BASE_URL}/Category/FilterProductBox?c={cate_id}&o=13&pi={next_page}",
                method="POST",
                body="IsParentCate=False&IsShowCompare=True&prevent=true",
                headers={
                    "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
                    "X-Requested-With": "XMLHttpRequest",
                    "Referer": referer,
                },
                callback=self.parse_listing,
                errback=self.handle_error,
                meta={"cate_id": cate_id, "page": next_page, "referer": referer},
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
            price_val = float(price_raw) if price_raw else None

            og_title = response.css("meta[property='og:title']::attr(content)").get() or ""
            name = re.split(r"\s+ra mắt|,", og_title)[0].strip() or None

            return {
                "name": name,
                "sku": None,  # no reliable SKU on campaign pages
                "brand": name.split()[0] if name else None,
                "price": price_val if price_val else None,
                "in_stock": False,  # price=0 means not yet on sale
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
        path_parts = [p for p in response.url.replace(BASE_URL, "").rstrip("/").split("/") if p]
        if len(path_parts) < 2:
            return
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

        if ld_product and not ld_product.get("sku") and not ld_product.get("url"):
            return

        product_url = response.url.split("?")[0]

        if not product_url:
            self.logger.error(f"Could not find URL for product at {response.url}")
            return

        item = ProductItem(url=product_url)

        item.source = "thegioididong"
        item.crawled_at = datetime.now(timezone.utc).isoformat()
        item.currency = offers.get("priceCurrency", "VND")
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

        # --- Fallback Path ---
        if not ld_product:
            gtm = self._parse_gtm_fallback(response)
            if not gtm:
                return
            item.name = gtm.get("name")
            item.sku = gtm.get("sku")
            item.description = None
            item.brand = gtm.get("brand") or ((gtm.get("name") or "").split()[0] or None)

            if ld_breadcrumbs:
                if len(ld_breadcrumbs) >= 3:
                    item.category = ld_breadcrumbs[-2].get("item", {}).get("name")
                    item.subcategory = ld_breadcrumbs[-1].get("item", {}).get("name")
                else:
                    item.category = ld_breadcrumbs[-1].get("item", {}).get("name")
                    item.subcategory = None

            item.price = gtm.get("price")
            item.in_stock = gtm.get("in_stock", False)
            if not product_images:
                og_image = response.css("meta[property='og:image']::attr(content)").get()
                product_images = [og_image.split("#")[0]] if og_image else []
            item.images = product_images

            item.specs = self._parse_gtm_specs(response)
            yield item
            return

        # --- Normal Path ---
        item.name = ld_product.get("name") or clean_text(response.css("h1::text").get())
        item.sku = ld_product.get("sku")

        desc_el = response.css("div.description div.text-detail")
        if desc_el:
            paragraphs = [
                re.sub(r"\s+", " ", strip_html(block)).strip()
                for block in desc_el.css("p, h2, h3, li").getall()
            ]
            item.description = "\n\n".join(p for p in paragraphs if p and p != "Xem thêm") or None
        else:
            item.description = ld_product.get("description")

        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        brand_match = re.search(r"\((.+?)\)", brand_raw) if brand_raw else None
        item.brand = brand_match.group(1) if brand_match else brand_raw

        if not item.brand and item.name:
            item.brand = item.name.split()[0] if item.name.split() else None

        if ld_breadcrumbs:
            if len(ld_breadcrumbs) >= 3:
                item.category = ld_breadcrumbs[-2].get("item", {}).get("name")
                item.subcategory = ld_breadcrumbs[-1].get("item", {}).get("name")
            else:
                item.category = ld_breadcrumbs[-1].get("item", {}).get("name")
                item.subcategory = None

        item.price = offers.get("price") or parse_price(
            response.css("[class*='price-present']::text").get()
        )
        item.original_price = parse_price(response.css("p.box-price-old::text").get())

        if item.original_price and item.price and item.original_price > item.price:
            item.discount_percent = round(
                ((item.original_price - item.price) / item.original_price) * 100
            )
        else:
            item.discount_percent = parse_discount(response.css("p.box-price-percent::text").get())

        availability = offers.get("availability", "")
        item.in_stock = "InStock" in availability or bool(
            response.css("[class*='btn-buynow']").get()
        )

        item.rating = rating.get("ratingValue") or parse_rating(
            response.css(".point-average-score::text").get()
        )
        item.review_count = rating.get("reviewcount")

        # Images
        ld_image = ld_product.get("image", {})
        ld_image_url = ld_image.get("contentUrl") if isinstance(ld_image, dict) else ld_image
        item.images = product_images or ([ld_image_url] if ld_image_url else [])

        item.specs = {
            prop["name"]: re.sub(
                r"\.\s*Xem thông tin hãng\s*$", "", strip_html(prop["value"])
            ).strip()
            for prop in (ld_product.get("additionalProperty") or [])
            if prop.get("name")
            and prop.get("value")
            and strip_html(prop["value"]).strip() != "Đang cập nhật"
        }

        yield item

    def handle_error(self, failure):
        if failure.check(HttpError):
            response = failure.value.response
            self.logger.error("HTTP %s for %s", response.status, failure.request.url)
        else:
            self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

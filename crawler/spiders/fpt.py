import json
from datetime import datetime, timezone

import psycopg2
import scrapy

from items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price

EXCLUDE_PATHS = {
    "/",
    "/gio-hang",
    "/tim-kiem",
    "/cua-hang",
    "/tos",
    "/sim-fpt",
    "/may-doi-tra",
}

FALLBACK_URLS = [
    "/dien-thoai",
    "/may-tinh-xach-tay",
    "/may-tinh-bang",
    "/may-tinh-de-ban",
    "/man-hinh",
    "/tivi",
    "/tu-lanh",
    "/may-giat",
    "/thiet-bi-bep",
    "/robot-hut-bui",
    "/may-loc-nuoc",
    "/may-loc-khong-khi",
    "/smartwatch",
    "/phu-kien",  # covers /tai-nghe, /loa, /may-anh etc. via subcategory discovery
]


class FptSpider(scrapy.Spider):
    name = "fpt"
    allowed_domains = ["fptshop.com.vn"]

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

    def parse_categories(self, response):
        all_links = response.css("a::attr(href)").getall()
        hrefs = [
            link
            for link in all_links
            if link.startswith("/")
            and link.count("/") == 1
            and link not in EXCLUDE_PATHS
            and not link.startswith("/tim-kiem")
        ]
        if not hrefs:
            self.logger.warning("Category discovery returned nothing — using fallback URLs")
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

    async def start(self):
        for href in FALLBACK_URLS:
            yield scrapy.Request(
                f"https://fptshop.com.vn{href}",
                callback=self.parse,
                errback=self.handle_error,
            )

        db_url = self.settings.get("DATABASE_URL")
        if db_url:
            try:
                conn = psycopg2.connect(db_url)
                cur = conn.cursor()
                cur.execute(
                    "SELECT url FROM products WHERE source = 'fpt' AND (price IS NULL OR in_stock = false)"
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

    def parse(self, response):
        seen = set()
        for href in response.css("[class*='ProductCard_cardDefault'] a::attr(href)").getall():
            if href in seen:
                continue
            seen.add(href)
            yield response.follow(href, callback=self.parse_product, errback=self.handle_error)
        base = response.url.rstrip("/").split("?")[0]
        path = base.replace("https://fptshop.com.vn", "")
        sub_seen = set()
        for href in response.css(f"a[href^='{path}/']::attr(href)").getall():
            if href.count("/") != 2:
                continue
            if href in sub_seen:
                continue
            sub_seen.add(href)
            yield response.follow(
                href,
                callback=self.parse,
                errback=self.handle_error,
            )

    def parse_product(self, response):
        item = ProductItem()
        item["source"] = "fpt"
        item["url"] = response.url
        item["crawled_at"] = datetime.now(timezone.utc).isoformat()
        item["currency"] = "VND"

        # Use stable script ID instead of looping all scripts
        product_data = {}
        raw = response.css("#detail-product-script::text").get()
        if raw:
            try:
                product_data = json.loads(raw)
            except (json.JSONDecodeError, AttributeError) as e:
                self.logger.warning("Failed to parse product JSON at %s: %s", response.url, e)
        # Category from breadcrumb JSON-LD (position 2 = top-level category e.g. "Điện thoại")
        category = None
        raw_bc = response.css("#breadcrumb-structured-data::text").get()
        if raw_bc:
            try:
                bc = json.loads(raw_bc)
                cat_item = next(
                    (x for x in bc.get("itemListElement", []) if x.get("position") == 2), None
                )
                if cat_item:
                    category = cat_item.get("name")
            except (json.JSONDecodeError, AttributeError) as e:
                self.logger.warning("Failed to parse breadcrumb JSON at %s: %s", response.url, e)
        if not category:
            url_parts = response.url.split("/")
            category = url_parts[3].replace("-", " ").title() if len(url_parts) > 3 else None
        item["category"] = category

        offers = product_data.get("offers", {})
        agg = product_data.get("aggregateRating", {})

        item["name"] = clean_text(product_data.get("name") or response.css("h1::text").get())
        item["price"] = parse_price(str(offers.get("price", "")))

        original = parse_price(response.css("span.line-through::text").get() or "")
        item["original_price"] = (
            original if original and item["price"] and original > item["price"] else None
        )

        if item["original_price"] and item["price"]:
            item["discount_percent"] = round(
                (item["original_price"] - item["price"]) / item["original_price"] * 100
            )
        else:
            # ['2', '%', '3.000.000đ'] → join first two → '2%'
            parts = response.css("span.text-red-red-7::text").getall()
            discount = parse_discount("".join(parts[:2])) if parts else None
            item["discount_percent"] = discount if discount and 0 < discount < 100 else None

        if item["price"] and item["discount_percent"] and not item["original_price"]:
            item["original_price"] = round(item["price"] / (1 - item["discount_percent"] / 100))

        item["brand"] = clean_text(product_data.get("brand", {}).get("name"))
        item["quantity"] = None
        item["in_stock"] = offers.get("availability", "").endswith("InStock")
        item["rating"] = float(agg["ratingValue"]) if agg.get("ratingValue") else None
        item["review_count"] = int(agg["reviewCount"]) if agg.get("reviewCount") else None

        # Single image string in SSR, wrap in list
        ld_image = product_data.get("image")
        item["images"] = (
            ld_image if isinstance(ld_image, list) else ([ld_image] if ld_image else [])
        )

        # Filter empty spec values (e.g. Chip: '')
        item["specs"] = {
            prop["name"]: prop["value"]
            for prop in (product_data.get("additionalProperty") or [])
            if prop.get("name") and str(prop.get("value", "")).strip()
        }

        yield item

    def handle_error(self, failure):
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

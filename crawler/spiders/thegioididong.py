import json
import re
from datetime import datetime, timezone

import psycopg2
import scrapy

from items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price, parse_rating

EXCLUDED_NAV = {"/may-in", "/muc-in", "/sim-so-dep"}

FALLBACK_URLS = [
    "/dtdd",
    "/laptop",
    "/may-tinh-bang",
    "/tai-nghe",
    "/phu-kien",
    "/dong-ho-thong-minh-ldp",
    "/dong-ho",
    "/may-doi-tra",
]


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
            for h in response.css("[class*='nav'] a::attr(href)").getall()
            if h and not h.startswith("javascript") and h not in EXCLUDED_NAV
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

        item["name"] = ld_product.get("name") or clean_text(response.css("h1::text").get())
        item["sku"] = ld_product.get("sku")
        item["description"] = ld_product.get("description")

        # Brand: {"@type": "Brand", "name": ["iPhone (Apple)"]} — name is a list
        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        brand_match = re.search(r"\((.+?)\)", brand_raw) if brand_raw else None
        item["brand"] = brand_match.group(1) if brand_match else brand_raw

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
        all_images = response.css("div.owl-carousel img::attr(src)").getall()
        product_images = list(dict.fromkeys(url for url in all_images if "/Products/" in url))
        item["images"] = product_images or ([ld_image_url] if ld_image_url else [])

        item["specs"] = {
            prop["name"]: prop["value"]
            for prop in (ld_product.get("additionalProperty") or [])
            if prop.get("name") and prop.get("value")
        }

        yield item

    def handle_error(self, failure):
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

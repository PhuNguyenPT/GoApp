from datetime import datetime, timezone

import psycopg2
import scrapy
from scrapy_playwright.page import PageMethod

from items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price


class FptSpider(scrapy.Spider):
    name = "fpt"
    allowed_domains = ["fptshop.com.vn"]
    start_urls = [
        "https://fptshop.com.vn/dien-thoai",
        "https://fptshop.com.vn/may-tinh-xach-tay",
        "https://fptshop.com.vn/may-tinh-bang",
        "https://fptshop.com.vn/phu-kien",
    ]

    custom_settings = {
        "DOWNLOAD_DELAY": 1.5,
        "RANDOMIZE_DOWNLOAD_DELAY": True,
        "DUPEFILTER_CLASS": "scrapy.dupefilters.RFPDupeFilter",
        "PLAYWRIGHT_DEFAULT_NAVIGATION_TIMEOUT": 60_000,
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

    _listing_meta = {
        "playwright": True,
        "playwright_page_goto_kwargs": {"wait_until": "domcontentloaded"},
        "playwright_page_methods": [
            PageMethod("evaluate", "window.scrollTo(0, document.body.scrollHeight)"),
            PageMethod("wait_for_timeout", 2000),
            PageMethod("evaluate", "window.scrollTo(0, document.body.scrollHeight)"),
            PageMethod("wait_for_timeout", 1000),
        ],
    }

    _product_meta = {
        "playwright": True,
        "playwright_page_goto_kwargs": {"wait_until": "domcontentloaded"},
        "playwright_include_page": False,
    }

    async def start(self):
        for url in self.start_urls:
            yield scrapy.Request(
                url,
                meta=self._listing_meta,
                callback=self.parse,
                errback=self.handle_error,
            )

        db_url = self.settings.get("DATABASE_URL")
        if db_url:
            try:
                conn = psycopg2.connect(db_url)
                cur = conn.cursor()
                cur.execute("SELECT url FROM products WHERE price IS NULL OR in_stock = false")
                for (url,) in cur.fetchall():
                    yield scrapy.Request(
                        url,
                        meta=self._product_meta,
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
            yield response.follow(
                href,
                callback=self.parse_product,
                errback=self.handle_error,
                meta=self._product_meta,
            )

    def parse_product(self, response):
        item = ProductItem()
        item["source"] = "fpt"
        item["url"] = response.url
        item["crawled_at"] = datetime.now(timezone.utc).isoformat()
        item["currency"] = "VND"
        item["name"] = clean_text(response.css("h1::text").get())

        import json

        json_ld = response.css("script[type='application/ld+json']::text").getall()
        product_data = {}
        for blob in json_ld:
            try:
                d = json.loads(blob)
                if d.get("@type") == "Product":
                    product_data = d
                    break
            except Exception:
                pass

        offers = product_data.get("offers", {})
        item["price"] = parse_price(str(offers.get("price", "")))
        original_raw = response.css("span.line-through::text").get() or ""
        original = parse_price(original_raw)
        item["original_price"] = (
            original if original and item["price"] and original > item["price"] else None
        )
        if item["original_price"] and item["price"]:
            diff = item["original_price"] - item["price"]
            percent = (diff / item["original_price"]) * 100
            item["discount_percent"] = round(percent)
        else:
            raw_discount = response.css("span.text-red-red-7::text").get()
            item["discount_percent"] = parse_discount(raw_discount) if raw_discount else None
        if item["price"] and item["discount_percent"] and not item["original_price"]:
            item["original_price"] = round(item["price"] / (1 - item["discount_percent"] / 100))
        item["brand"] = clean_text(
            product_data.get("brand", {}).get("name")
            or response.css("[class*='brand'] img::attr(alt)").get()
        )
        item["category"] = (
            clean_text(response.css("ol li:nth-last-child(2) a::text").get())
            or response.url.split("/")[3].replace("-", " ").title()
        )
        item["in_stock"] = offers.get("availability", "").endswith("InStock")
        agg = product_data.get("aggregateRating", {})
        item["rating"] = float(agg["ratingValue"]) if agg.get("ratingValue") else None
        item["review_count"] = int(agg["reviewCount"]) if agg.get("reviewCount") else None
        images = product_data.get("image", [])
        item["images"] = images if isinstance(images, list) else [images]
        item["specs"] = {}
        spec_table = response.css("table.technical-content tr, [class*='Specification'] tr")
        for row in spec_table:
            key = clean_text(row.css("td:first-child::text").get())
            val = clean_text(row.css("td:last-child::text").get())
            if key and val:
                item["specs"][key] = val

        for prop in product_data.get("additionalProperty", []):
            name = prop.get("name")
            value = prop.get("value")
            if name and value:
                item["specs"][name] = value
        yield item

    def handle_error(self, failure):
        from playwright._impl._errors import TimeoutError as PlaywrightTimeout

        if failure.check(PlaywrightTimeout):
            request = failure.request
            retries = request.meta.get("_timeout_retries", 0)
            max_retries = 3
            if retries < max_retries:
                self.logger.warning(
                    "Timeout on %s — retrying (%d/%d)", request.url, retries + 1, max_retries
                )
                new_request = request.copy()
                new_request.meta["_timeout_retries"] = retries + 1
                new_request.dont_filter = True
                yield new_request
                return

        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

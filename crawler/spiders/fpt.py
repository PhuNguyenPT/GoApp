import re
import scrapy
from datetime import datetime, timezone
from items import ProductItem
from utils.helpers import parse_price, parse_discount, clean_text, parse_rating


class FptSpider(scrapy.Spider):
    name = "fpt"
    allowed_domains = ["fptshop.com.vn"]
    start_urls = [
        "https://fptshop.com.vn/dien-thoai",
        "https://fptshop.com.vn/may-tinh-xach-tay",
        "https://fptshop.com.vn/may-tinh-bang",
        "https://fptshop.com.vn/phu-kien"
    ]

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
        for url in self.start_urls:
            yield scrapy.Request(
                url,
                meta={"playwright": True},
                callback=self.parse,
                errback=self.handle_error,
            )
                    
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
                meta={"playwright": True, "playwright_include_page": False},
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
        item["original_price"] = parse_price(
            response.css("span.line-through::text").get() or ""
        )
        item["discount_percent"] = parse_discount(
            response.css("span.text-red-red-7::text").get() or ""
        )
        item["brand"] = clean_text(product_data.get("brand", {}).get("name") or
                                response.css("[class*='brand'] img::attr(alt)").get())
        item["category"] = clean_text(response.css("ol li:nth-last-child(2) a::text").get()) or \
                   response.url.split("/")[3].replace("-", " ").title()
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
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))
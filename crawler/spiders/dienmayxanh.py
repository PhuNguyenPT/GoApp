import re
from datetime import datetime, timezone

import scrapy

from crawler.items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price, parse_rating


class DienmayxanhSpider(scrapy.Spider):
    name = "dienmayxanh"
    allowed_domains = ["dienmayxanh.com"]
    start_urls = [
        "https://www.dienmayxanh.com/tivi",
        "https://www.dienmayxanh.com/tu-lanh",
        "https://www.dienmayxanh.com/may-giat",
        "https://www.dienmayxanh.com/dieu-hoa",
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

    def parse(self, response):
        for link in response.css("ul.listproduct li.item a.main-contain::attr(href)").getall():
            yield response.follow(link, callback=self.parse_product, errback=self.handle_error)
        next_page = response.css("a.next::attr(href)").get()
        if next_page:
            yield response.follow(next_page, callback=self.parse, errback=self.handle_error)

    def parse_product(self, response):
        item = ProductItem()
        item["source"] = "dienmayxanh"
        item["url"] = response.url
        item["crawled_at"] = datetime.now(timezone.utc).isoformat()
        item["currency"] = "VND"
        item["name"] = clean_text(response.css("h1.product-name::text").get())
        item["brand"] = clean_text(response.css("div.parameter a[href*='/hang/']::text").get())
        item["category"] = clean_text(
            response.css("ol.breadcrumb li:nth-last-child(2) a::text").get()
        )
        item["price"] = parse_price(response.css("p.box-price-present strong::text").get())
        item["original_price"] = parse_price(response.css("p.box-price-old::text").get())
        item["discount_percent"] = parse_discount(response.css("p.box-price-percent::text").get())
        item["in_stock"] = bool(
            response.css("button.btn-buy, button[class*='add-cart'], div.box-buy button")
        )
        item["rating"] = parse_rating(response.css("div.rank-point span::text").get())
        item["review_count"] = None
        review_text = response.css("a.rating-count::text").get()
        if review_text:
            match = re.search(r"\d+", review_text)
            item["review_count"] = int(match.group()) if match else None
        item["images"] = (
            response.css("div.product-gallery img::attr(data-src)").getall()
            or response.css("div.product-gallery img::attr(src)").getall()
        )
        specs = {}
        for row in response.css("table.parameter tr"):
            key = clean_text(" ".join(row.css("td:first-child ::text").getall()))
            val = clean_text(" ".join(row.css("td:last-child ::text").getall()))
            if key and val:
                specs[key] = val
        item["specs"] = specs
        yield item

    def handle_error(self, failure):
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

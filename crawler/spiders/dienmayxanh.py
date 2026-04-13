import json
import re
from datetime import datetime, timezone

import scrapy

from items import ProductItem
from utils.helpers import clean_text, parse_price, parse_rating

EXCLUDED_NAV = {
    "flashsale",
    "online-only",
    "khuyen-mai",
    "cart",
    "may-doi-tra",
    "he-thong-sieu-thi",
    "tra-gop",
    "bao-hanh",
    "tho-dien-may-xanh",
    "dich-vu-bao-duong",
    "kinh-nghiem-hay",
    "hang-cao-cap",
    "chuong-trinh",
    "thong-tin",
    "gioi-thieu",
    "lien-he",
    "hoi-dap",
    "tin-tuc",
    "phieu-mua-hang",
    "b2b",
    "danh-muc-nhom-hang",
}

FALLBACK_URLS = [
    "/tivi",
    "/may-lanh",
    "/tu-lanh",
    "/may-giat",
    "/may-say-quan-ao",
    "/may-nuoc-nong",
    "/tu-dong",
    "/may-rua-chen",
    "/loa-ldp",
    "/gia-dung",
    "/may-loc-nuoc",
    "/noi-com-dien",
    "/quat-dieu-hoa",
    "/quat",
    "/may-hut-bui",
    "/robot-hut-bui",
    "/may-loc-khong-khi",
    "/may-xay-sinh-to",
    "/noi-chien-khong-dau",
    "/bep-tu",
    "/bep-hong-ngoai",
    "/bep-ga",
    "/lo-vi-song",
    "/binh-dun-sieu-toc",
    "/may-say-toc",
    "/may-cao-rau",
    "/may-massage-toan-than",
    "/ghe-massage",
    "/dien-thoai",
    "/laptop",
    "/may-tinh-bang",
    "/dong-ho-thong-minh",
    "/may-tinh-nguyen-bo",
    "/man-hinh-may-tinh",
    "/may-in",
    "/phu-kien",
    "/tai-nghe",
    "/sac-dtdd",
    "/chuot-may-tinh",
    "/ban-phim",
    "/camera-giam-sat",
    "/dong-ho-deo-tay",
    "/may-khoan",
    "/may-bom-nuoc",
]


def strip_html(text):
    return re.sub(r"<[^>]+>", "", text).strip() if text else text


class DienmayxanhSpider(scrapy.Spider):
    name = "dienmayxanh"
    allowed_domains = ["dienmayxanh.com"]

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
            "Referer": "https://www.dienmayxanh.com/",
        },
    }

    async def start(self):
        yield scrapy.Request(
            "https://www.dienmayxanh.com",
            callback=self.parse_categories,
            errback=self.handle_error,
        )

    def parse_categories(self, response):
        hrefs = []
        for h in response.css("[class*='main'] a::attr(href)").getall():
            if not h:
                continue
            if not h.startswith("/"):  # skip bare slugs and external URLs
                continue
            if h.startswith("javascript"):
                continue
            if "?" in h or "#" in h:
                continue
            parts = h.lstrip("/").split("/")
            if len(parts) > 1:  # skip product-level paths e.g. /may-lanh/product-slug
                continue
            slug = parts[0]
            if any(slug.startswith(ex) for ex in EXCLUDED_NAV):
                continue
            hrefs.append(h)

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
                link.split("?")[0],
                callback=self.parse_product,
                errback=self.handle_error,
            )
        next_page = response.css("a.next::attr(href)").get()
        if next_page:
            yield response.follow(next_page, callback=self.parse, errback=self.handle_error)

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
        item["source"] = "dienmayxanh"
        item["url"] = ld_product.get("url") or response.url.split("?")[0]
        item["crawled_at"] = datetime.now(timezone.utc).isoformat()
        item["currency"] = offers.get("priceCurrency", "VND")

        item["name"] = ld_product.get("name") or clean_text(response.css("h1::text").get())
        item["sku"] = ld_product.get("sku")
        item["description"] = ld_product.get("description")

        # Brand: name may be a list e.g. ["Sony"]
        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        item["brand"] = brand_raw or (item["name"].split()[0] if item["name"] else None)

        # Category from BreadcrumbList (skip position 1 = site name)
        if ld_breadcrumbs:
            if len(ld_breadcrumbs) >= 3:
                item["category"] = ld_breadcrumbs[-2].get("item", {}).get("name")
                item["subcategory"] = ld_breadcrumbs[-1].get("item", {}).get("name")
            else:
                item["category"] = ld_breadcrumbs[-1].get("item", {}).get("name")
                item["subcategory"] = None
        else:
            breadcrumbs = response.css("[class*='breadcrumb'] a::text").getall()
            if len(breadcrumbs) >= 2:
                item["category"] = clean_text(breadcrumbs[-2])
                item["subcategory"] = clean_text(breadcrumbs[-1])
            else:
                item["category"] = clean_text(breadcrumbs[-1]) if breadcrumbs else None
                item["subcategory"] = None

        item["price"] = parse_price(
            response.css("input#DisPriceScenrioGTM::attr(value)").get()
        ) or offers.get("price")
        item["original_price"] = parse_price(
            response.css("input#PriceOriginGTM::attr(value)").get()
        )
        if item["original_price"] and item["original_price"] == item["price"]:
            item["original_price"] = None

        if item["original_price"] and item["price"] and item["original_price"] > item["price"]:
            diff = item["original_price"] - item["price"]
            item["discount_percent"] = round((diff / item["original_price"]) * 100)
        else:
            discount_raw = response.css("input#PercentScenrioGTM::attr(value)").get()
            item["discount_percent"] = (
                int(discount_raw) if discount_raw and discount_raw.strip() != "0" else None
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

        # aggregateRating uses lowercase "reviewcount" on dienmayxanh
        item["rating"] = rating.get("ratingValue") or parse_rating(
            response.css("[class*='rank'] span::text").get()
        )
        item["review_count"] = rating.get("reviewcount") or rating.get("reviewCount")

        # Images: prefer data-src (lazy-loaded), filter to /Products/ URLs only
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
            ld_image = ld_product.get("image", {})
            ld_image_url = ld_image.get("contentUrl") if isinstance(ld_image, dict) else ld_image
            product_images = [ld_image_url] if ld_image_url else []
        item["images"] = product_images

        # Specs from additionalProperty, strip HTML and skip placeholder values
        item["specs"] = {
            prop["name"]: re.sub(
                r"\.\s*Xem thông tin hãng\s*$", "", strip_html(prop["value"])
            ).strip()
            for prop in (ld_product.get("additionalProperty") or [])
            if prop.get("name")
            and prop.get("value")
            and strip_html(prop["value"]).strip() not in ("Đang cập nhật", "Hãng không công bố")
        }

        yield item

    def handle_error(self, failure):
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

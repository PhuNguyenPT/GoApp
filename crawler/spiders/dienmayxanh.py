import json
import re
from datetime import datetime, timezone

import scrapy

from items import ProductItem
from utils.helpers import clean_text, parse_discount, parse_price, parse_rating

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

PAGE_SIZE = 20

API_HEADERS = {
    "Content-Type": "application/x-www-form-urlencoded; charset=UTF-8",
    "X-Requested-With": "XMLHttpRequest",
}

UNAVAILABLE_STATUSES = ("Tin đồn", "Ngừng kinh doanh", "Hết hàng", "Không kinh doanh")


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
        start_url = getattr(self, "start_url", None)
        if start_url:
            # Product URLs have two path segments: /category/product-slug
            path = start_url.rstrip("/").replace("https://www.dienmayxanh.com", "")
            parts = [p for p in path.split("/") if p]
            if len(parts) >= 2:
                # Looks like a product page — go direct
                yield scrapy.Request(
                    start_url,
                    callback=self.parse_product,
                    errback=self.handle_error,
                )
            else:
                # Looks like a category page
                yield scrapy.Request(
                    start_url,
                    callback=self.parse_category_page,
                    errback=self.handle_error,
                )
        else:
            yield scrapy.Request(
                "https://www.dienmayxanh.com",
                callback=self.parse_categories,
                errback=self.handle_error,
            )

    def parse_categories(self, response):
        hrefs = []
        for h in response.css("[class*='main'] a::attr(href)").getall():
            if not h or not h.startswith("/"):
                continue
            if h.startswith("javascript"):
                continue
            if "?" in h or "#" in h:
                continue
            parts = h.lstrip("/").split("/")
            if len(parts) > 1:
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
                callback=self.parse_category_page,
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
            self.logger.debug("No cate_id found at %s — skipping", response.url)
            return

        self.logger.debug("Discovered cate_id=%s for %s", cate_id, response.url)

        yield scrapy.Request(
            f"https://www.dienmayxanh.com/Category/FilterProductBox?c={cate_id}&o=13&pi=1",
            method="POST",
            body="IsParentCate=false&prevent=true",
            headers={**API_HEADERS, "Referer": response.url},
            callback=self.parse_listing,
            errback=self.handle_error,
            meta={"cate_id": cate_id, "page": 1, "referer": response.url},
        )

    def parse_listing(self, response):
        data = response.json()
        total = data.get("total", 0)
        html = data.get("listproducts", "")
        sel = scrapy.Selector(text=html)

        for a in sel.css("li.item a.main-contain"):
            href = a.attrib.get("href", "").split("?")[0]
            if not href:
                continue
            yield response.follow(
                href,
                callback=self.parse_product,
                errback=self.handle_error,
                meta={"product_status": a.attrib.get("data-productstatus", "").strip()},
            )

        page = response.meta["page"]
        cate_id = response.meta["cate_id"]
        referer = response.meta["referer"]

        if page * PAGE_SIZE < total:
            next_page = page + 1
            yield scrapy.Request(
                f"https://www.dienmayxanh.com/Category/FilterProductBox?c={cate_id}&o=13&pi={next_page}",
                method="POST",
                body="IsParentCate=false&prevent=true",
                headers={**API_HEADERS, "Referer": referer},
                callback=self.parse_listing,
                errback=self.handle_error,
                meta={"cate_id": cate_id, "page": next_page, "referer": referer},
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
        agg_rating = ld_product.get("aggregateRating") or {}

        # --- name ---
        name = ld_product.get("name") or clean_text(response.css("h1::text").get())

        # --- brand ---
        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        brand = brand_raw or (name.split()[0] if name else None)

        # --- category / subcategory ---
        if ld_breadcrumbs:
            if len(ld_breadcrumbs) >= 3:
                category = ld_breadcrumbs[-2].get("item", {}).get("name")
                subcategory = ld_breadcrumbs[-1].get("item", {}).get("name")
            else:
                category = ld_breadcrumbs[-1].get("item", {}).get("name")
                subcategory = None
        else:
            breadcrumbs = response.css("[class*='breadcrumb'] a::text").getall()
            if len(breadcrumbs) >= 2:
                category = clean_text(breadcrumbs[-2])
                subcategory = clean_text(breadcrumbs[-1])
            else:
                category = clean_text(breadcrumbs[-1]) if breadcrumbs else None
                subcategory = None

        # --- description ---
        desc_els = response.css("div#tab-2 div.text-detail p, div#tab-2 div.text-detail h3")
        if desc_els:
            parts = [re.sub(r"\s+", " ", strip_html(el.get())).strip() for el in desc_els]
            description = "\n\n".join(p for p in parts if p) or None
        else:
            description = ld_product.get("description")

        # --- pricing ---
        _raw_price = parse_price(response.css("input#DisPriceScenrioGTM::attr(value)").get())
        _ld_price = offers.get("price")

        price = (_raw_price if _raw_price and _raw_price > 0 else None) or (
            _ld_price if _ld_price and float(_ld_price) > 0 else None
        )

        original_price = parse_price(response.css("input#PriceOriginGTM::attr(value)").get())
        if original_price and original_price == price:
            original_price = None

        if original_price and price and original_price > price:
            discount_percent = round(((original_price - price) / original_price) * 100)
        else:
            discount_percent = parse_discount(
                response.css("input#PercentScenrioGTM::attr(value)").get()
            )

        if price and discount_percent and not original_price:
            original_price = (
                round(price / (1 - discount_percent / 100)) if 0 < discount_percent < 100 else None
            )

        # --- images ---
        sku = ld_product.get("sku")

        src_images = [
            url
            for url in response.css("div.owl-carousel img::attr(src)").getall()
            if "/Products/" in url
            and ("/Slider/" in url or "/Images/" in url)
            and sku
            and f"/{sku}/" in url
        ]
        data_src_images = [
            url
            for url in response.css("div.owl-carousel img::attr(data-src)").getall()
            if "/Products/" in url
            and ("/Slider/" in url or "/Images/" in url)
            and sku
            and f"/{sku}/" in url
        ]
        images = list(dict.fromkeys(src_images + data_src_images))
        if not images:
            ld_image = ld_product.get("image", {})
            ld_image_url = ld_image.get("contentUrl") if isinstance(ld_image, dict) else ld_image
            images = [ld_image_url] if ld_image_url else []

        # --- specs ---
        specs = {
            prop["name"]: re.sub(
                r"\.\s*Xem thông tin hãng\s*$", "", strip_html(prop["value"])
            ).strip()
            for prop in (ld_product.get("additionalProperty") or [])
            if prop.get("name")
            and prop.get("value")
            and strip_html(prop["value"]).strip() not in ("Đang cập nhật", "Hãng không công bố")
        }

        # --- in_stock ---
        product_status = (
            response.meta.get("product_status")
            or response.css("strong.productstatus::text").get("").strip()
        )

        in_stock = not any(
            s in product_status for s in UNAVAILABLE_STATUSES
        ) and "InStock" in offers.get("availability", "")

        yield ProductItem(
            url=ld_product.get("url") or response.url.split("?")[0],
            source="dienmayxanh",
            crawled_at=datetime.now(timezone.utc),
            currency=offers.get("priceCurrency", "VND"),
            name=name,
            sku=ld_product.get("sku"),
            brand=brand,
            category=category,
            subcategory=subcategory,
            description=description,
            price=price,
            original_price=original_price,
            discount_percent=discount_percent,
            in_stock=in_stock,
            quantity=None,
            rating=agg_rating.get("ratingValue")
            or parse_rating(response.css("[class*='rank'] span::text").get()),
            review_count=agg_rating.get("reviewcount") or agg_rating.get("reviewCount"),
            images=images,
            specs=specs,
        )

    def handle_error(self, failure):
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

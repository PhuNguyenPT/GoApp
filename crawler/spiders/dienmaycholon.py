import json
import re
from datetime import datetime, timezone

import scrapy
from scrapy.spidermiddlewares.httperror import HttpError

from items import ProductItem
from utils.helpers import parse_price, parse_rating

CDN = "https://cdn11.dienmaycholon.vn/filewebdmclnew/DMCL21/Picture"
BASE_URL = "https://dienmaycholon.com"

API_HEADERS = {
    "Accept": "application/json, text/plain, */*",
    "X-Requested-With": "XMLHttpRequest",
    "Referer": f"{BASE_URL}/",
}
PAGE_SIZE = 15

EXCLUDED_SLUGS = {
    "flashsale",
    "khuyen-mai",
    "cart",
    "tra-gop",
    "bao-hanh",
    "he-thong-sieu-thi",
    "gioi-thieu",
    "lien-he",
    "tin-tuc",
    "chinh-sach",
    "dich-vu",
    "chuong-trinh",
    "thong-tin",
    "hang-cao-cap",
    "b2b",
    "hoi-dap",
    "phieu-mua-hang",
    "gio-hang",
    "livestream",
    "trang-",
    "hop-tac-mat-bang",
    "mua-hang-doanh-nghiep",
    "mo-the-vpbank-nhan-voucher",
}


def strip_html(text):
    return re.sub(r"<[^>]+>", "", text).strip() if text else text


class DienmaycholonSpider(scrapy.Spider):
    name = "dienmaycholon"
    allowed_domains = ["dienmaycholon.com"]
    SEO_JUNK = re.compile(
        r"chính hãng.*giá tốt|đặt hàng ngay|ưu đãi hấp dẫn|giao hàng nhanh"
        r"|Siêu Thị Điện Máy.*Chợ Lớn",
        re.IGNORECASE,
    )

    async def start(self):
        start_url = getattr(self, "start_url", None)
        if start_url:
            path = start_url.rstrip("/").replace(BASE_URL, "")
            parts = [p for p in path.split("/") if p]
            if len(parts) >= 2:
                yield scrapy.Request(
                    start_url, callback=self.parse_product, errback=self.handle_error
                )
            else:
                yield scrapy.Request(
                    start_url, callback=self.parse_category_page, errback=self.handle_error
                )
            return
        yield scrapy.Request(BASE_URL, callback=self.parse_categories, errback=self.handle_error)

    def parse_categories(self, response):
        seen = set()
        for href in response.css("a::attr(href)").getall():
            if (
                not href
                or href.startswith("tel:")
                or href.startswith("mailto:")
                or href.startswith("javascript:")
            ):
                continue
            path = href.replace(BASE_URL, "").strip("/")
            if not path or "/" in path or "?" in path or "#" in path:
                continue
            if any(path.startswith(ex) for ex in EXCLUDED_SLUGS):
                continue
            if path in seen:
                continue
            seen.add(path)
            yield response.follow(
                f"/{path}",
                callback=self.parse_category_page,
                errback=self.handle_error,
            )

    def parse_category_page(self, response):
        cate_id = response.css(".my_cate::attr(idx)").get()
        if not cate_id:
            ids = re.findall(r'idx=["\'](\d+)["\']', response.text)
            cate_id = ids[0] if ids else None
        if not cate_id:
            self.logger.debug("No cate_id found at %s — skipping", response.url)
            return

        alias = response.url.rstrip("/").split("/")[-1]

        # Read brand filter directly from the page's series attribute
        series = response.css(".my_cate::attr(series)").get()
        brand_filter = f"&s[]={series}" if series else ""

        self.logger.debug("Discovered cate_id=%s alias=%s series=%s", cate_id, alias, series)

        yield scrapy.Request(
            f"{BASE_URL}/api/product/cate?page=1&offset=0&id={cate_id}{brand_filter}",
            headers={**API_HEADERS, "Referer": response.url},
            callback=self.parse_listing,
            errback=self.handle_error,
            meta={
                "alias": alias,
                "cate_id": int(cate_id),
                "brand_filter": brand_filter,
                "cate_url": response.url,
            },
        )

    def parse_listing(self, response):
        data = response.json()["data"]
        total_pages = data.get("Totalpage") or 0
        current_page = data.get("getCurrentPageNumber") or 1
        alias = response.meta["alias"]
        cate_id = response.meta["cate_id"]
        brand_filter = response.meta.get("brand_filter", "")
        cate_url = response.meta.get("cate_url", f"{BASE_URL}/{alias}")

        for product in data.get("data") or []:
            if not isinstance(product, dict):
                if product:
                    self.logger.warning(
                        "Unexpected product type %s at %s: %r", type(product), response.url, product
                    )
                continue
            alias_val = product.get("alias")
            if not alias_val:
                continue
            yield scrapy.Request(
                f"{BASE_URL}/{alias}/{alias_val}",
                callback=self.parse_product,
                errback=self.handle_error,
                meta={"api_data": product},
                headers={"Referer": cate_url},
            )

        if current_page == 1:
            for page in range(2, total_pages + 1):
                offset = (page - 1) * PAGE_SIZE
                yield scrapy.Request(
                    f"{BASE_URL}/api/product/cate"
                    f"?page={page}&offset={offset}&id={cate_id}{brand_filter}",
                    headers={**API_HEADERS, "Referer": cate_url},
                    callback=self.parse_listing,
                    errback=self.handle_error,
                    meta={
                        "alias": alias,
                        "cate_id": cate_id,
                        "brand_filter": brand_filter,
                        "cate_url": cate_url,
                    },
                )

    def parse_product(self, response):
        api = response.meta.get("api_data", {})
        flag_content = api.get("flag_content")
        if not isinstance(flag_content, dict):
            flag_content = {}

        ld_product = {}
        ld_breadcrumbs = []
        for script in response.css("script[type='application/ld+json']::text").getall():
            try:
                cleaned = re.sub(r"[\x00-\x1f\x7f]", " ", script)
                d = json.loads(cleaned)
                if d.get("@type") == "Product":
                    ld_product = d
                elif d.get("@type") == "BreadcrumbList":
                    ld_breadcrumbs = d.get("itemListElement", [])
            except (json.JSONDecodeError, AttributeError) as e:
                self.logger.warning("LD+JSON parse error at %s: %s", response.url, e)

        offers = ld_product.get("offers") or {}
        agg_rating = ld_product.get("aggregateRating") or {}

        # --- brand ---
        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        series = flag_content.get("series", [])
        brand = brand_raw or (series[0] if series else None)

        # --- category / subcategory ---
        crumbs = [e for e in ld_breadcrumbs if e.get("position", 0) > 1]
        if len(crumbs) >= 2:
            category = crumbs[-2].get("item", {}).get("name")
            subcategory = crumbs[-1].get("item", {}).get("name")
        elif crumbs:
            category = crumbs[-1].get("item", {}).get("name")
            subcategory = None
        else:
            category = subcategory = None

        # --- description ---

        desc_els = (
            response.css(
                "div.tab_feature p:not([style*='center']):not(.see_feature),"
                "div.tab_feature h2,"
                "div.tab_feature h3"
            )
            or response.css("div.detail-content p, div.detail-content h2, div.detail-content h3")
            or response.css("div#tab-description p")
            or response.css(
                "div.des_pro_item p:not([style*='center']),div.des_pro_item h2,div.des_pro_item h3"
            )
        )
        if desc_els:
            parts = [
                re.sub(r"\s+", " ", strip_html(el.get())).strip()
                for el in desc_els
                if not self.SEO_JUNK.search(el.get())  # ← filter at element level
            ]
            description = "\n\n".join(p for p in parts if p) or None
        else:
            description = None

        # Discard LD+JSON/meta SEO placeholder descriptions
        if not description:
            ld_desc = ld_product.get("description")
            if ld_desc and not self.SEO_JUNK.search(ld_desc):
                description = ld_desc

        # --- pricing ---
        price = parse_price(api.get("discount")) or parse_price(
            response.css("strong.price_sale::text").get()
        )
        original_price = parse_price(api.get("saleprice")) or parse_price(
            response.css("div.price_giaban span::text").get()
        )

        if original_price and original_price == price:
            original_price = None

        if original_price and price and original_price > price:
            discount_percent = round(((original_price - price) / original_price) * 100)
        else:
            discount_raw = response.css("span.discount_percent::text").get()
            if discount_raw:
                m = re.search(r"\d+", discount_raw)
                discount_percent = int(m.group()) if m else None
            else:
                discount_percent = None

        # --- images ---
        images = list(
            dict.fromkeys(
                u
                for u in response.css("div.box_pro-images img::attr(data-src)").getall()
                if not u.startswith("data:image")
            )
        )
        if not images:
            ld_images = ld_product.get("image", [])
            images = [ld_images] if isinstance(ld_images, str) else (ld_images or [])

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

        yield ProductItem(
            url=response.url.split("?")[0],
            source="dienmaycholon",
            crawled_at=datetime.now(timezone.utc),
            currency="VND",
            name=ld_product.get("name") or api.get("name"),
            sku=ld_product.get("sku") or api.get("sap_code"),
            brand=brand,
            category=category,
            subcategory=subcategory,
            description=description,
            price=price,
            original_price=original_price,
            discount_percent=discount_percent,
            in_stock=(
                "InStock" in offers.get("availability", "")
                or bool(response.css("button.click_buy").get())
            ),
            quantity=None,
            rating=agg_rating.get("ratingValue")
            or flag_content.get("rate_customer")
            or parse_rating(response.css("[class*='rating']::text").get()),
            review_count=agg_rating.get("reviewcount")
            or agg_rating.get("reviewCount")
            or flag_content.get("totalrate_customer"),
            images=images,
            specs=specs,
        )

    def handle_error(self, failure):
        if failure.check(HttpError):
            response = failure.value.response
            if response.status == 404:
                self.logger.debug("404 skipped: %s", failure.request.url)
                return
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

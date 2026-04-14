import json
import re
from datetime import datetime, timezone

import scrapy

from items import ProductItem
from utils.helpers import parse_price, parse_rating

CDN = "https://cdn11.dienmaycholon.vn/filewebdmclnew/DMCL21/Picture"

API_HEADERS = {
    "Accept": "application/json, text/plain, */*",
    "X-Requested-With": "XMLHttpRequest",
    "Referer": "https://dienmaycholon.com/",
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


class DienmaycholonSpider(scrapy.Spider):
    name = "dienmaycholon"
    allowed_domains = ["dienmaycholon.com"]

    async def start(self):
        yield scrapy.Request(
            "https://dienmaycholon.com",
            callback=self.parse_categories,
            errback=self.handle_error,
        )

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
            path = href.replace("https://dienmaycholon.com", "").strip("/")
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
        self.logger.debug("Discovered cate_id=%s for alias=%s", cate_id, alias)

        yield scrapy.Request(
            f"https://dienmaycholon.com/api/product/cate?page=1&offset=0&id={cate_id}",
            headers={**API_HEADERS, "Referer": response.url},
            callback=self.parse_listing,
            errback=self.handle_error,
            meta={"alias": alias, "cate_id": int(cate_id)},
        )

    def parse_listing(self, response):
        data = response.json()["data"]
        total_pages = data.get("Totalpage") or 0
        current_page = data.get("getCurrentPageNumber") or 1
        alias = response.meta["alias"]
        cate_id = response.meta["cate_id"]

        for product in data.get("data") or []:
            if not isinstance(product, dict):
                self.logger.warning(
                    "Unexpected product type %s at %s: %r", type(product), response.url, product
                )
                continue
            alias_val = product.get("alias")
            if not alias_val:
                continue
            yield scrapy.Request(
                f"https://dienmaycholon.com/{alias}/{alias_val}",
                callback=self.parse_product,
                errback=self.handle_error,
                meta={"api_data": product},
            )

        if current_page == 1:
            for page in range(2, total_pages + 1):
                offset = (page - 1) * PAGE_SIZE
                yield scrapy.Request(
                    f"https://dienmaycholon.com/api/product/cate"
                    f"?page={page}&offset={offset}&id={cate_id}",
                    headers={**API_HEADERS, "Referer": f"https://dienmaycholon.com/{alias}"},
                    callback=self.parse_listing,
                    errback=self.handle_error,
                    meta={"alias": alias, "cate_id": cate_id},
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

        item = ProductItem()
        item["source"] = "dienmaycholon"
        item["url"] = response.url.split("?")[0]
        item["crawled_at"] = datetime.now(timezone.utc).isoformat()
        item["currency"] = "VND"

        item["name"] = ld_product.get("name") or api.get("name")
        item["sku"] = ld_product.get("sku") or api.get("sap_code")
        item["description"] = ld_product.get("description")

        brand_raw = ld_product.get("brand", {}).get("name")
        if isinstance(brand_raw, list):
            brand_raw = brand_raw[0] if brand_raw else None
        series = flag_content.get("series", [])
        item["brand"] = brand_raw or (series[0] if series else None)

        crumbs = [e for e in ld_breadcrumbs if e.get("position", 0) > 1]
        if len(crumbs) >= 2:
            item["category"] = crumbs[-2].get("item", {}).get("name")
            item["subcategory"] = crumbs[-1].get("item", {}).get("name")
        elif crumbs:
            item["category"] = crumbs[-1].get("item", {}).get("name")
            item["subcategory"] = None
        else:
            item["category"] = None
            item["subcategory"] = None

        # Prices from CSS (confirmed selectors) with API fallback
        item["price"] = parse_price(response.css("strong.price_sale::text").get()) or parse_price(
            str(api.get("discount", ""))
        )

        item["original_price"] = parse_price(
            response.css("div.price_giaban span::text").get()
        ) or parse_price(str(api.get("saleprice", "")))

        if item["original_price"] and item["original_price"] == item["price"]:
            item["original_price"] = None

        if item["original_price"] and item["price"] and item["original_price"] > item["price"]:
            diff = item["original_price"] - item["price"]
            item["discount_percent"] = round((diff / item["original_price"]) * 100)
        else:
            discount_raw = response.css("span.discount_percent::text").get()
            if discount_raw:
                m = re.search(r"\d+", discount_raw)
                item["discount_percent"] = int(m.group()) if m else None
            else:
                item["discount_percent"] = None

        availability = offers.get("availability", "")
        item["in_stock"] = "InStock" in availability or bool(response.css("button.click_buy").get())
        item["quantity"] = None

        item["rating"] = (
            agg_rating.get("ratingValue")
            or flag_content.get("rate_customer")
            or parse_rating(response.css("[class*='rating']::text").get())
        )
        item["review_count"] = (
            agg_rating.get("reviewcount")
            or agg_rating.get("reviewCount")
            or flag_content.get("totalrate_customer")
        )

        data_src = [
            u
            for u in response.css("[class*='product'] img::attr(data-src)").getall()
            if "data:image" not in u and "/Apro/Apro_product_" in u
        ]
        src = [
            u
            for u in response.css("[class*='product'] img::attr(src)").getall()
            if "data:image" not in u and "/Apro/Apro_product_" in u
        ]
        product_images = list(dict.fromkeys(data_src + src))
        if not product_images:
            ld_images = ld_product.get("image", [])
            product_images = [ld_images] if isinstance(ld_images, str) else (ld_images or [])
        item["images"] = product_images

        def strip_html(text):
            return re.sub(r"<[^>]+>", "", text).strip() if text else text

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
        from scrapy.spidermiddlewares.httperror import HttpError

        if failure.check(HttpError):
            response = failure.value.response
            if response.status == 404:
                self.logger.debug("404 skipped: %s", failure.request.url)
                return
        self.logger.error("Request failed: %s — %s", failure.request.url, repr(failure))

import scrapy


class ProductItem(scrapy.Item):
    id = scrapy.Field()
    url = scrapy.Field()
    source = scrapy.Field()
    name = scrapy.Field()
    brand = scrapy.Field()
    category = scrapy.Field()
    subcategory = scrapy.Field()
    description = scrapy.Field()
    sku = scrapy.Field()
    images = scrapy.Field()
    specs = scrapy.Field()
    currency = scrapy.Field()
    price = scrapy.Field()
    original_price = scrapy.Field()
    discount_percent = scrapy.Field()
    quantity = scrapy.Field()
    in_stock = scrapy.Field()
    rating = scrapy.Field()
    review_count = scrapy.Field()
    crawled_at = scrapy.Field()

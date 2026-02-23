import scrapy


class ProductItem(scrapy.Item):
    name = scrapy.Field()
    url = scrapy.Field()
    source = scrapy.Field()
    price = scrapy.Field()
    original_price = scrapy.Field()
    discount_percent = scrapy.Field()
    currency = scrapy.Field()
    sku = scrapy.Field()
    brand = scrapy.Field()
    category = scrapy.Field()
    description = scrapy.Field()
    rating = scrapy.Field()
    review_count = scrapy.Field()
    in_stock = scrapy.Field()
    images = scrapy.Field()
    specs = scrapy.Field()
    crawled_at = scrapy.Field()

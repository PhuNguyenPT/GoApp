import json
import logging
import os
from datetime import datetime

import psycopg2
from itemadapter import ItemAdapter

logger = logging.getLogger(__name__)


class JsonWriterPipeline:
    @classmethod
    def from_crawler(cls, crawler):
        instance = cls()
        instance.settings = crawler.settings
        instance.crawler = crawler
        return instance

    def open_spider(self):
        name = self.crawler.spider.name if self.crawler.spider else "unknown"
        os.makedirs(self.settings["OUTPUT_DIR"], exist_ok=True)
        fname = f"{self.settings['OUTPUT_DIR']}/{name}_{datetime.now():%Y%m%d_%H%M%S}.jsonl"
        self.file = open(fname, "w", encoding="utf-8")

    def close_spider(self):
        if getattr(self, "file", None):
            self.file.close()

    def process_item(self, item):
        line = json.dumps(ItemAdapter(item).asdict(), ensure_ascii=False)
        self.file.write(line + "\n")
        return item


class PostgresPipeline:
    BATCH_SIZE = 50

    @classmethod
    def from_crawler(cls, crawler):
        instance = cls()
        instance.settings = crawler.settings
        return instance

    def open_spider(self):
        db_url = self.settings.get("DATABASE_URL")
        if not db_url:
            logger.warning("DATABASE_URL not set — skipping Postgres pipeline")
            self.conn = None
            return
        self.conn = psycopg2.connect(db_url)
        self.cur = self.conn.cursor()
        self._item_count = 0

    def close_spider(self):
        if not getattr(self, "conn", None):
            return
        self.conn.commit()
        self.cur.close()
        self.conn.close()

    def process_item(self, item):
        if not getattr(self, "conn", None):
            return item
        data = ItemAdapter(item).asdict()
        self.cur.execute(
            """
            INSERT INTO products (
                name, url, source, price, original_price, discount_percent,
                currency, brand, category, rating, review_count,
                in_stock, images, specs, crawled_at
            ) VALUES (
                %(name)s, %(url)s, %(source)s, %(price)s, %(original_price)s,
                %(discount_percent)s, %(currency)s, %(brand)s, %(category)s,
                %(rating)s, %(review_count)s, %(in_stock)s, %(images)s,
                %(specs)s, %(crawled_at)s
            )
            ON CONFLICT (url) DO UPDATE SET
                name = EXCLUDED.name,
                price = EXCLUDED.price,
                original_price = EXCLUDED.original_price,
                discount_percent = EXCLUDED.discount_percent,
                in_stock = EXCLUDED.in_stock,
                rating = EXCLUDED.rating,
                review_count = EXCLUDED.review_count,
                images = EXCLUDED.images,
                crawled_at = EXCLUDED.crawled_at
            """,
            {
                **data,
                "images": json.dumps(data.get("images", [])),
                "specs": json.dumps(data.get("specs", {})),
            },
        )
        self._item_count += 1
        if self._item_count % self.BATCH_SIZE == 0:
            self.conn.commit()
        return item

import json
import logging
import os
from dataclasses import asdict
from datetime import datetime

import psycopg2

from db import save_item
from items import ProductItem

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

    def process_item(self, item: ProductItem):

        line = json.dumps(asdict(item), ensure_ascii=False, default=str)
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

    def process_item(self, item: ProductItem):
        if not getattr(self, "conn", None):
            return item

        try:
            save_item(self.cur, item)
            self._item_count += 1
            if self._item_count % self.BATCH_SIZE == 0:
                self.conn.commit()
        except Exception as e:
            self.conn.rollback()
            self.cur = self.conn.cursor()
            logger.warning(f"Failed to save item (url={item.url}): {e}")

        return item

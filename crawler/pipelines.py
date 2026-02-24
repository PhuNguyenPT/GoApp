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
        line = json.dumps(ItemAdapter(item).asdict(), ensure_ascii=False, default=str)
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

        product_id = self._upsert_product(data)

        source_id = self._upsert_source(data)

        dim_product_id = self._upsert_dim_product(product_id, data)

        date_id = self._get_date_id(data.get("crawled_at"))

        if all([dim_product_id, source_id, date_id]):
            self._insert_fact_snapshot(dim_product_id, source_id, date_id, data)

        self._item_count += 1
        if self._item_count % self.BATCH_SIZE == 0:
            self.conn.commit()

        return item

    def _upsert_product(self, data):
        self.cur.execute(
            """
            INSERT INTO products (
                url, source, name, brand, category,
                price, original_price, currency,
                in_stock, quantity, rating, review_count,
                images, specs, crawled_at
            ) VALUES (
                %(url)s, %(source)s, %(name)s, %(brand)s, %(category)s,
                %(price)s, %(original_price)s, %(currency)s,
                %(in_stock)s, %(quantity)s, %(rating)s, %(review_count)s,
                %(images)s, %(specs)s, %(crawled_at)s
            )
            ON CONFLICT (url) DO UPDATE SET
                name             = EXCLUDED.name,
                price            = EXCLUDED.price,
                original_price   = EXCLUDED.original_price,
                in_stock         = EXCLUDED.in_stock,
                quantity         = EXCLUDED.quantity,
                rating           = EXCLUDED.rating,
                review_count     = EXCLUDED.review_count,
                images           = EXCLUDED.images,
                specs            = EXCLUDED.specs,
                crawled_at       = EXCLUDED.crawled_at
            RETURNING id
            """,
            {
                **data,
                "images": json.dumps(data.get("images", [])),
                "specs": json.dumps(data.get("specs", {})),
            },
        )
        return self.cur.fetchone()[0]

    def _upsert_source(self, data):
        source = data.get("source") or "unknown"
        domain = self._source_domain(source)
        self.cur.execute(
            """
            INSERT INTO warehouse.dim_source (name, domain)
            VALUES (%s, %s)
            ON CONFLICT (name) DO NOTHING
            RETURNING id
            """,
            (source, domain),
        )
        row = self.cur.fetchone()
        if row:
            return row[0]
        self.cur.execute("SELECT id FROM warehouse.dim_source WHERE name = %s", (source,))
        return self.cur.fetchone()[0]

    def _upsert_dim_product(self, product_id, data):
        self.cur.execute(
            """
            SELECT id, name, brand, category
            FROM warehouse.dim_product
            WHERE product_id = %s AND is_current = TRUE
            """,
            (product_id,),
        )
        current = self.cur.fetchone()

        scd_fields = ("name", "brand", "category")
        changed = current and any(data.get(f) != current[i + 1] for i, f in enumerate(scd_fields))

        if current and not changed:
            return current[0]

        if changed:
            self.cur.execute(
                """
                UPDATE warehouse.dim_product
                SET valid_to = NOW(), is_current = FALSE
                WHERE product_id = %s AND is_current = TRUE
                """,
                (product_id,),
            )

        self.cur.execute(
            """
            INSERT INTO warehouse.dim_product (
                product_id, url, name, brand, category,
                currency, images, specs
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            RETURNING id
            """,
            (
                product_id,
                data.get("url"),
                data.get("name"),
                data.get("brand"),
                data.get("category"),
                data.get("currency"),
                json.dumps(data.get("images", [])),
                json.dumps(data.get("specs", {})),
            ),
        )
        return self.cur.fetchone()[0]

    def _get_date_id(self, crawled_at):
        if not crawled_at:
            return None
        dt = datetime.fromisoformat(crawled_at).date()
        self.cur.execute(
            "SELECT date_id FROM warehouse.dim_date WHERE full_date = %s",
            (dt,),
        )
        row = self.cur.fetchone()
        return row[0] if row else None

    def _insert_fact_snapshot(self, dim_product_id, source_id, date_id, data):
        self.cur.execute(
            """
            INSERT INTO warehouse.fact_product_snapshot (
                dim_product_id, source_id, date_id,
                price, original_price, discount_percent,
                quantity, in_stock, rating, review_count,
                crawled_at
            ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """,
            (
                dim_product_id,
                source_id,
                date_id,
                data.get("price"),
                data.get("original_price"),
                data.get("discount_percent"),
                data.get("quantity"),
                data.get("in_stock"),
                data.get("rating"),
                data.get("review_count"),
                data.get("crawled_at"),
            ),
        )

    @staticmethod
    def _source_domain(source):
        return {
            "dienmaycholon": "dienmaycholon.com",
            "dienmayxanh": "dienmayxanh.com",
            "fpt": "fptshop.com.vn",
            "thegioididong": "thegioididong.com",
        }.get(source)

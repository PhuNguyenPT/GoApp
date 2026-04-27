import json
from datetime import datetime

from items import ProductItem

SOURCE_DOMAIN = {
    "dienmaycholon": "dienmaycholon.com",
    "dienmayxanh": "dienmayxanh.com",
    "fptshop": "fptshop.com.vn",
    "thegioididong": "thegioididong.com",
}


def upsert_product(cur, item: ProductItem):
    cur.execute(
        """
        INSERT INTO products (
            url, source, sku, name, brand, category, subcategory,
            description, price, original_price, discount_percent, currency,
            in_stock, quantity, rating, review_count,
            images, specs, crawled_at
        ) VALUES (
            %(url)s, %(source)s, %(sku)s, %(name)s, %(brand)s, %(category)s, %(subcategory)s,
            %(description)s, %(price)s, %(original_price)s, %(discount_percent)s, %(currency)s,
            %(in_stock)s, %(quantity)s, %(rating)s, %(review_count)s,
            %(images)s, %(specs)s, %(crawled_at)s
        )
        ON CONFLICT (url) DO UPDATE SET
            sku              = EXCLUDED.sku,
            name             = EXCLUDED.name,
            brand            = EXCLUDED.brand,
            category         = EXCLUDED.category,
            subcategory      = EXCLUDED.subcategory,
            description      = EXCLUDED.description,
            price            = EXCLUDED.price,
            original_price   = EXCLUDED.original_price,
            discount_percent = EXCLUDED.discount_percent,
            currency         = EXCLUDED.currency,
            in_stock         = EXCLUDED.in_stock,
            quantity         = EXCLUDED.quantity,
            rating           = EXCLUDED.rating,
            review_count     = EXCLUDED.review_count,
            images           = EXCLUDED.images,
            specs            = EXCLUDED.specs,
            crawled_at       = EXCLUDED.crawled_at,
            updated_at       = NOW()
        RETURNING id
        """,
        {
            "url": item.url,
            "source": item.source,
            "sku": item.sku,
            "name": item.name,
            "brand": item.brand,
            "category": item.category,
            "subcategory": item.subcategory,
            "description": item.description,
            "price": item.price,
            "original_price": item.original_price,
            "discount_percent": item.discount_percent,
            "currency": item.currency,
            "in_stock": item.in_stock,
            "quantity": item.quantity,
            "rating": item.rating,
            "review_count": item.review_count,
            "images": json.dumps(item.images),
            "specs": json.dumps(item.specs),
            "crawled_at": item.crawled_at,
        },
    )
    return cur.fetchone()[0]


def insert_product_history(cur, product_id, item: ProductItem):
    cur.execute(
        """
        INSERT INTO products_history (
            product_id, url, source, name, brand, category, subcategory,
            price, original_price, discount_percent, currency,
            in_stock, quantity, rating, review_count,
            images, specs, crawled_at
        ) VALUES (
            %(product_id)s, %(url)s, %(source)s, %(name)s, %(brand)s, %(category)s, %(subcategory)s,
            %(price)s, %(original_price)s, %(discount_percent)s, %(currency)s,
            %(in_stock)s, %(quantity)s, %(rating)s, %(review_count)s,
            %(images)s, %(specs)s, %(crawled_at)s
        )
        """,
        {
            "product_id": product_id,
            "url": item.url,
            "source": item.source,
            "name": item.name,
            "brand": item.brand,
            "category": item.category,
            "subcategory": item.subcategory,
            "price": item.price,
            "original_price": item.original_price,
            "discount_percent": item.discount_percent,
            "currency": item.currency,
            "in_stock": item.in_stock,
            "quantity": item.quantity,
            "rating": item.rating,
            "review_count": item.review_count,
            "images": json.dumps(item.images),
            "specs": json.dumps(item.specs),
            "crawled_at": item.crawled_at,
        },
    )


def upsert_source(cur, item: ProductItem):
    source = item.source or "unknown"
    domain = SOURCE_DOMAIN.get(source)
    cur.execute(
        """
        INSERT INTO warehouse.dim_source (name, domain)
        VALUES (%s, %s)
        ON CONFLICT (name) DO NOTHING
        RETURNING id
        """,
        (source, domain),
    )
    row = cur.fetchone()
    if row:
        return row[0]
    cur.execute("SELECT id FROM warehouse.dim_source WHERE name = %s", (source,))
    return cur.fetchone()[0]


def upsert_dim_product(cur, product_id, item: ProductItem):
    cur.execute(
        """
        SELECT id, name, brand, category, subcategory
        FROM warehouse.dim_product
        WHERE product_id = %s AND is_current = TRUE
        """,
        (product_id,),
    )
    current = cur.fetchone()
    scd_fields = ("name", "brand", "category", "subcategory")
    changed = current and any(getattr(item, f) != current[i + 1] for i, f in enumerate(scd_fields))

    if current and not changed:
        return current[0]

    if changed:
        cur.execute(
            """
            UPDATE warehouse.dim_product
            SET valid_to = NOW(), is_current = FALSE
            WHERE product_id = %s AND is_current = TRUE
            """,
            (product_id,),
        )

    cur.execute(
        """
        INSERT INTO warehouse.dim_product (
            product_id, url, name, brand, category, subcategory,
            currency, images, specs
        ) VALUES (
            %(product_id)s, %(url)s, %(name)s, %(brand)s, %(category)s, %(subcategory)s,
            %(currency)s, %(images)s, %(specs)s
        )
        RETURNING id
        """,
        {
            "product_id": product_id,
            "url": item.url,
            "name": item.name,
            "brand": item.brand,
            "category": item.category,
            "subcategory": item.subcategory,
            "currency": item.currency,
            "images": json.dumps(item.images),
            "specs": json.dumps(item.specs),
        },
    )
    return cur.fetchone()[0]


def get_date_id(cur, crawled_at):
    if not crawled_at:
        return None
    dt = (
        crawled_at.date()
        if hasattr(crawled_at, "date")
        else datetime.fromisoformat(crawled_at).date()
    )
    cur.execute("SELECT date_id FROM warehouse.dim_date WHERE full_date = %s", (dt,))
    row = cur.fetchone()
    return row[0] if row else None


def insert_fact_snapshot(cur, dim_product_id, source_id, date_id, item: ProductItem):
    cur.execute(
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
            item.price,
            item.original_price,
            item.discount_percent,
            item.quantity,
            item.in_stock,
            item.rating,
            item.review_count,
            item.crawled_at,
        ),
    )


def save_item(cur, item: ProductItem):
    product_id = upsert_product(cur, item)
    insert_product_history(cur, product_id, item)
    source_id = upsert_source(cur, item)
    dim_product_id = upsert_dim_product(cur, product_id, item)
    date_id = get_date_id(cur, item.crawled_at)

    if all([dim_product_id, source_id, date_id]):
        insert_fact_snapshot(cur, dim_product_id, source_id, date_id, item)
        return True
    return False

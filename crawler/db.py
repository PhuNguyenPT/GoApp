import json
from datetime import datetime

SOURCE_DOMAIN = {
    "dienmaycholon": "dienmaycholon.com",
    "dienmayxanh": "dienmayxanh.com",
    "fptshop": "fptshop.com.vn",
    "thegioididong": "thegioididong.com",
}


def upsert_product(cur, data):
    cur.execute(
        """
        INSERT INTO products (
            url, source, sku, name, brand, category,
            description, price, original_price, discount_percent, currency,
            in_stock, quantity, rating, review_count,
            images, specs, crawled_at
        ) VALUES (
            %(url)s, %(source)s, %(sku)s, %(name)s, %(brand)s, %(category)s,
            %(description)s, %(price)s, %(original_price)s, %(discount_percent)s, %(currency)s,
            %(in_stock)s, %(quantity)s, %(rating)s, %(review_count)s,
            %(images)s, %(specs)s, %(crawled_at)s
        )
        ON CONFLICT (url) DO UPDATE SET
            sku              = EXCLUDED.sku,
            name             = EXCLUDED.name,
            brand            = EXCLUDED.brand,
            category         = EXCLUDED.category,
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
            **data,
            "sku": data.get("sku"),
            "description": data.get("description"),
            "images": json.dumps(data.get("images", [])),
            "specs": json.dumps(data.get("specs", {})),
        },
    )
    return cur.fetchone()[0]


def upsert_source(cur, data):
    source = data.get("source") or "unknown"
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


def upsert_dim_product(cur, product_id, data):
    cur.execute(
        """
        SELECT id, name, brand, category
        FROM warehouse.dim_product
        WHERE product_id = %s AND is_current = TRUE
        """,
        (product_id,),
    )
    current = cur.fetchone()
    scd_fields = ("name", "brand", "category")
    changed = current and any(data.get(f) != current[i + 1] for i, f in enumerate(scd_fields))

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
    return cur.fetchone()[0]


def get_date_id(cur, crawled_at):
    if not crawled_at:
        return None
    dt = datetime.fromisoformat(crawled_at).date()
    cur.execute("SELECT date_id FROM warehouse.dim_date WHERE full_date = %s", (dt,))
    row = cur.fetchone()
    return row[0] if row else None


def insert_fact_snapshot(cur, dim_product_id, source_id, date_id, data):
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


def save_item(cur, data):
    """Run the full insert sequence for one item. Returns True if fact snapshot was inserted."""
    product_id = upsert_product(cur, data)
    source_id = upsert_source(cur, data)
    dim_product_id = upsert_dim_product(cur, product_id, data)
    date_id = get_date_id(cur, data.get("crawled_at"))

    if all([dim_product_id, source_id, date_id]):
        insert_fact_snapshot(cur, dim_product_id, source_id, date_id, data)
        return True
    return False

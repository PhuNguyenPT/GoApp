-- +goose Up
-- +goose StatementBegin
CREATE SCHEMA IF NOT EXISTS warehouse;

CREATE TABLE IF NOT EXISTS products_history (
    history_id       UUID        PRIMARY KEY DEFAULT uuidv7(),
    product_id       UUID        NOT NULL,
    url              TEXT        NOT NULL,
    source           TEXT,
    name             TEXT,
    brand            TEXT,
    category         TEXT,
    price            NUMERIC,
    original_price   NUMERIC,
    discount_percent NUMERIC,
    currency         TEXT,
    in_stock         BOOLEAN,
    quantity         INT,
    rating           NUMERIC,
    review_count     INT,
    images           JSONB,
    specs            JSONB,
    crawled_at       TIMESTAMPTZ,
    changed_at       TIMESTAMPTZ      NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION record_product_history()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO products_history (
        product_id, url, source, name, brand, category,
        price, original_price, discount_percent, currency,
        in_stock, quantity, rating, review_count,
        images, specs, crawled_at
    ) VALUES (
        OLD.id, OLD.url, OLD.source, OLD.name, OLD.brand, OLD.category,
        OLD.price, OLD.original_price, OLD.discount_percent, OLD.currency,
        OLD.in_stock, OLD.quantity, OLD.rating, OLD.review_count,
        OLD.images, OLD.specs, OLD.crawled_at
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_products_history
BEFORE UPDATE ON products
FOR EACH ROW EXECUTE FUNCTION record_product_history();

CREATE TABLE IF NOT EXISTS warehouse.dim_date (
    date_id     INT     PRIMARY KEY,  -- YYYYMMDD
    full_date   DATE    NOT NULL,
    year        INT     NOT NULL,
    quarter     INT     NOT NULL,
    month       INT     NOT NULL,
    month_name  TEXT    NOT NULL,
    week        INT     NOT NULL,
    day_of_week INT     NOT NULL,
    day_name    TEXT    NOT NULL,
    is_weekend  BOOLEAN NOT NULL
);

INSERT INTO warehouse.dim_date (date_id, full_date, year, quarter, month, month_name, week, day_of_week, day_name, is_weekend)
SELECT
    TO_CHAR(d, 'YYYYMMDD')::INT,
    d,
    EXTRACT(YEAR    FROM d)::INT,
    EXTRACT(QUARTER FROM d)::INT,
    EXTRACT(MONTH   FROM d)::INT,
    TO_CHAR(d, 'Month'),
    EXTRACT(WEEK    FROM d)::INT,
    EXTRACT(ISODOW  FROM d)::INT,
    TO_CHAR(d, 'Day'),
    EXTRACT(ISODOW  FROM d) IN (6, 7)
FROM generate_series('2020-01-01'::DATE, '2035-12-31'::DATE, '1 day') AS d;

CREATE TABLE IF NOT EXISTS warehouse.dim_source (
    id         UUID        PRIMARY KEY DEFAULT uuidv7(),
    name       TEXT        UNIQUE NOT NULL,
    domain     TEXT,
    country    TEXT        NOT NULL DEFAULT 'VN',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);


CREATE TABLE IF NOT EXISTS warehouse.dim_product (
    id         UUID        PRIMARY KEY DEFAULT uuidv7(),
    product_id UUID        NOT NULL REFERENCES products(id),
    url        TEXT        NOT NULL,
    name       TEXT,
    brand      TEXT,
    category   TEXT,
    currency   TEXT,
    images     JSONB,
    specs      JSONB,
    valid_from TIMESTAMPTZ      NOT NULL DEFAULT NOW(),
    valid_to   TIMESTAMPTZ,
    is_current BOOLEAN          NOT NULL DEFAULT TRUE
);

CREATE TABLE IF NOT EXISTS warehouse.fact_product_snapshot (
    id             UUID        PRIMARY KEY DEFAULT uuidv7(),
    dim_product_id UUID        NOT NULL REFERENCES warehouse.dim_product(id),
    source_id      UUID        NOT NULL REFERENCES warehouse.dim_source(id),
    date_id        INT         NOT NULL REFERENCES warehouse.dim_date(date_id),
    price          NUMERIC,
    original_price NUMERIC,
    discount_percent NUMERIC,
    quantity       INT,
    in_stock       BOOLEAN,
    rating         NUMERIC,
    review_count   INT,
    crawled_at     TIMESTAMPTZ NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE    IF EXISTS warehouse.fact_product_snapshot;
DROP TABLE    IF EXISTS warehouse.dim_product;
DROP TABLE    IF EXISTS warehouse.dim_source;
DROP TABLE    IF EXISTS warehouse.dim_date;
DROP TRIGGER  IF EXISTS trg_products_history ON products;
DROP FUNCTION IF EXISTS record_product_history;
DROP TABLE    IF EXISTS warehouse.products_history;
DROP SCHEMA IF EXISTS warehouse;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
    id SERIAL PRIMARY KEY,
    url TEXT UNIQUE NOT NULL,
    source TEXT,
    name TEXT,
    brand TEXT,
    category TEXT,
    price NUMERIC,
    original_price NUMERIC,
    discount_percent INT,
    currency TEXT DEFAULT 'VND',
    in_stock BOOLEAN,
    rating NUMERIC,
    review_count INT,
    images JSONB,
    specs JSONB,
    crawled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
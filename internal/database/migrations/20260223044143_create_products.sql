-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    url TEXT UNIQUE NOT NULL,
    source TEXT,
    sku TEXT,
    name TEXT,
    brand TEXT,
    category TEXT,
    subcategory TEXT,
    description TEXT,
    price NUMERIC,
    original_price NUMERIC,
    discount_percent INT,
    currency TEXT DEFAULT 'VND',
    in_stock BOOLEAN,
    quantity INT,
    rating NUMERIC,
    review_count INT,
    images JSONB,
    specs JSONB,
    crawled_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS products;
-- +goose StatementEnd
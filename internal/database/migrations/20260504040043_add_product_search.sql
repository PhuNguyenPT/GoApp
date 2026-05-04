-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE EXTENSION IF NOT EXISTS unaccent;

CREATE TEXT SEARCH CONFIGURATION vietnamese (COPY = simple);
ALTER TEXT SEARCH CONFIGURATION vietnamese
    ALTER MAPPING FOR hword, hword_part, word WITH unaccent, simple;

ALTER TABLE products
ADD COLUMN IF NOT EXISTS search_vector tsvector
GENERATED ALWAYS AS (
    setweight(to_tsvector('vietnamese', coalesce(name, '')), 'A') ||
    setweight(to_tsvector('vietnamese', coalesce(brand, '')), 'B') ||
    setweight(to_tsvector('vietnamese', coalesce(category, '')), 'C') ||
    setweight(to_tsvector('vietnamese', coalesce(subcategory, '')), 'C') ||
    setweight(to_tsvector('vietnamese', coalesce(sku, '')), 'D')
) STORED;

CREATE INDEX IF NOT EXISTS idx_products_search_fts  ON products USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_products_search_trgm ON products USING GIN(name gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE products DROP COLUMN IF EXISTS search_vector;
DROP INDEX IF EXISTS idx_products_search_fts;
DROP INDEX IF EXISTS idx_products_search_trgm;
DROP TEXT SEARCH CONFIGURATION IF EXISTS vietnamese;
DROP EXTENSION IF EXISTS unaccent;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd
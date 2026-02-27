-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_products_source_category ON products (lower(source), lower(category), crawled_at DESC);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_products_source_category;
-- +goose StatementEnd
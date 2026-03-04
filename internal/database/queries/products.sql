-- name: GetDistinctSources :many
SELECT DISTINCT source FROM products
WHERE source IS NOT NULL
ORDER BY source;

-- name: GetCategoriesBySource :many
SELECT DISTINCT category FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND category IS NOT NULL
ORDER BY category;

-- name: GetProductsBySourceAndCategory :many
SELECT * FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text))
ORDER BY crawled_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountProductsBySourceAndCategory :one
SELECT COUNT(*) FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text));

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1;
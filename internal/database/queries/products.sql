-- name: GetDistinctSources :many
SELECT DISTINCT source FROM products
WHERE source IS NOT NULL
ORDER BY source;

-- name: GetCategoriesBySource :many
SELECT DISTINCT category FROM products
WHERE ($1 = '' OR lower(source) = lower($1))
AND category IS NOT NULL
ORDER BY category;

-- name: GetProductsBySourceAndCategory :many
SELECT * FROM products
WHERE ($1 = '' OR lower(source) = lower($1))
AND ($2 = '' OR lower(category) = lower($2))
ORDER BY crawled_at DESC
LIMIT $3 OFFSET $4;

-- name: CountProductsBySourceAndCategory :one
SELECT COUNT(*) FROM products
WHERE ($1 = '' OR lower(source) = lower($1))
AND ($2 = '' OR lower(category) = lower($2));

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1;
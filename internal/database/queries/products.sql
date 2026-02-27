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
SELECT id, url, source, name, brand, category, price, original_price,
       discount_percent, currency, in_stock, rating, review_count,
       images, crawled_at, created_at
FROM products
WHERE ($1 = '' OR lower(source) = lower($1))
AND ($2 = '' OR lower(category) = lower($2))
ORDER BY crawled_at DESC
LIMIT $3 OFFSET $4;

-- name: CountProductsBySourceAndCategory :one
SELECT COUNT(*) FROM products
WHERE ($1 = '' OR lower(source) = lower($1))
AND ($2 = '' OR lower(category) = lower($2));

-- name: GetProductByID :one
SELECT id, url, source, name, brand, category, price, original_price,
       discount_percent, currency, in_stock, rating, review_count,
       images, specs, crawled_at, created_at
FROM products
WHERE id = $1;
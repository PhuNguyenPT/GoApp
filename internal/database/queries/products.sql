-- name: GetDistinctSources :many
SELECT DISTINCT source FROM products
WHERE source IS NOT NULL
ORDER BY source;

-- name: GetCategoriesBySource :many
SELECT DISTINCT category FROM products
WHERE (source = $1 OR $1 = '')
AND category IS NOT NULL
ORDER BY category;

-- name: GetProductsBySourceAndCategory :many
SELECT id, url, source, name, brand, category, price, original_price,
       discount_percent, currency, in_stock, rating, review_count,
       images, crawled_at, created_at
FROM products
WHERE (source = $1 OR $1 = '')
AND (category = $2 OR $2 = '')
ORDER BY created_at DESC
LIMIT $3 OFFSET $4;

-- name: CountProductsBySourceAndCategory :one
SELECT COUNT(*) FROM products
WHERE (source = $1 OR $1 = '')
AND (category = $2 OR $2 = '');
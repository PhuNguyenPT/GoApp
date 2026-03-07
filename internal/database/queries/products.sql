-- name: GetDistinctSources :many
SELECT DISTINCT source FROM products
WHERE source IS NOT NULL
ORDER BY source;

-- name: GetCategoriesBySource :many
SELECT DISTINCT category FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND category IS NOT NULL
ORDER BY category;

-- name: GetSubcategoriesBySourceAndCategory :many
SELECT DISTINCT subcategory FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text))
AND subcategory IS NOT NULL
ORDER BY subcategory;

-- name: GetProductsBySourceAndCategory :many
SELECT * FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text))
AND (sqlc.arg(subcategory)::text = '' OR lower(subcategory) = lower(sqlc.arg(subcategory)::text))
ORDER BY crawled_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: CountProductsBySourceAndCategory :one
SELECT COUNT(*) FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text))
AND (sqlc.arg(subcategory)::text = '' OR lower(subcategory) = lower(sqlc.arg(subcategory)::text))
AND (sqlc.arg(min_price)::numeric = 0 OR price >= sqlc.arg(min_price)::numeric)
AND (sqlc.arg(max_price)::numeric = 0 OR price <= sqlc.arg(max_price)::numeric);

-- name: GetProductByID :one
SELECT * FROM products
WHERE id = $1;

-- name: GetProductSummaries :many
SELECT id, source, name, brand, category,
       price, original_price, discount_percent, currency,
       in_stock, quantity, rating, review_count, images
FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text))
AND (sqlc.arg(subcategory)::text = '' OR lower(subcategory) = lower(sqlc.arg(subcategory)::text))
AND (sqlc.arg(min_price)::numeric = 0 OR price >= sqlc.arg(min_price)::numeric)
AND (sqlc.arg(max_price)::numeric = 0 OR price <= sqlc.arg(max_price)::numeric)
ORDER BY
  CASE WHEN sqlc.arg(sort_field)::text = 'price::numeric'   AND sqlc.arg(sort_dir)::text = 'asc'  THEN price::numeric END ASC,
  CASE WHEN sqlc.arg(sort_field)::text = 'price::numeric'   AND sqlc.arg(sort_dir)::text = 'desc' THEN price::numeric END DESC,
  CASE WHEN sqlc.arg(sort_field)::text = 'rating::numeric'  AND sqlc.arg(sort_dir)::text = 'asc'  THEN rating::numeric END ASC,
  CASE WHEN sqlc.arg(sort_field)::text = 'rating::numeric'  AND sqlc.arg(sort_dir)::text = 'desc' THEN rating::numeric END DESC,
  CASE WHEN sqlc.arg(sort_field)::text = 'name'             AND sqlc.arg(sort_dir)::text = 'asc'  THEN name END ASC,
  CASE WHEN sqlc.arg(sort_field)::text = 'name'             AND sqlc.arg(sort_dir)::text = 'desc' THEN name END DESC,
  CASE WHEN sqlc.arg(sort_field)::text = 'crawled_at'       AND sqlc.arg(sort_dir)::text = 'asc'  THEN crawled_at END ASC,
  CASE WHEN sqlc.arg(sort_field)::text = 'crawled_at'       AND sqlc.arg(sort_dir)::text = 'desc' THEN crawled_at END DESC,
  crawled_at DESC
LIMIT sqlc.arg(page_limit) OFFSET sqlc.arg(page_offset);

-- name: GetPricePercentiles :one
SELECT
    PERCENTILE_DISC(0.25) WITHIN GROUP (ORDER BY price::numeric)::text AS p25,
    PERCENTILE_DISC(0.50) WITHIN GROUP (ORDER BY price::numeric)::text AS p50,
    PERCENTILE_DISC(0.75) WITHIN GROUP (ORDER BY price::numeric)::text AS p75,
    MIN(price::numeric)::text AS min_price,
    MAX(price::numeric)::text AS max_price,
    MODE() WITHIN GROUP (ORDER BY currency)::text AS currency
FROM products
WHERE (sqlc.arg(source)::text = '' OR lower(source) = lower(sqlc.arg(source)::text))
AND (sqlc.arg(category)::text = '' OR lower(category) = lower(sqlc.arg(category)::text))
AND (sqlc.arg(subcategory)::text = '' OR lower(subcategory) = lower(sqlc.arg(subcategory)::text))
AND price IS NOT NULL;
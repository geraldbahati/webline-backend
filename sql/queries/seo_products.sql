-- name: UpdateProductSEO :exec
UPDATE products
SET
    part_number = $2,
    meta_title = $3,
    meta_description = $4,
    meta_keywords = $5,
    updated_at = now()
WHERE id = $1;

-- name: GetProductSEO :one
WITH rate AS (
    SELECT COALESCE(
        (
            SELECT rate_to_kes
            FROM exchange_rates
            WHERE currency_code = 'USD'
              AND (valid_to IS NULL OR valid_to >= NOW())
              AND valid_from <= NOW()
            ORDER BY valid_from DESC
            LIMIT 1
        ),
        135
    ) AS rate_to_kes
)
SELECT
    p.id,
    p.part_number,
    p.meta_title,
    p.meta_description,
    p.meta_keywords,
    p.usd_price,
    (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
    c.name AS brand_name,
    COALESCE(
        (
            SELECT image_url
            FROM product_images
            WHERE product_id = p.id
            ORDER BY created_at
            LIMIT 1
        ),
        ''
    )::TEXT AS image_url
FROM products p
JOIN categories c ON p.category_id = c.id
CROSS JOIN rate
WHERE p.slug = $1;


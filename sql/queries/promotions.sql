-- name: CreatePromotion :one
INSERT INTO promotions ( title, description, image_url, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, title, description, image_url, start_date, end_date, created_at, updated_at;

-- name: GetPromotions :many
SELECT
    p.title AS name,
    p.slug AS slug,
    p.description AS description,
    p.image_url AS imageUrl,
    CAST(
            CASE
                WHEN COUNT(pp.product_id) = 1 THEN (
                    SELECT pr.slug
                    FROM promotion_products pp_inner
                             JOIN products pr ON pp_inner.product_id = pr.id
                    WHERE pp_inner.promotion_id = p.id
                    LIMIT 1
                )
                ELSE ''
                END AS text
    ) AS productSlug
FROM
    promotions p
        LEFT JOIN
    promotion_products pp ON p.id = pp.promotion_id
GROUP BY
    p.id, p.title, p.slug, p.description, p.image_url, p.created_at
ORDER BY
    p.created_at;


-- name: GetV2Promotions :many
SELECT
    p.id,
    p.title AS name,
    p.slug,
    CASE
        WHEN COUNT(pp.product_id) = 1 THEN 'feature'
        ELSE 'sale'
        END AS type,
    p.image_url,
    COUNT(pp.product_id) AS numberOfProducts,
    p.status,
    p.start_date AS startDate,
    p.end_date AS endDate
FROM
    promotions p
        LEFT JOIN
    promotion_products pp ON p.id = pp.promotion_id
        LEFT JOIN
    discounts d ON pp.product_id = d.product_id
GROUP BY
    p.id, p.title, p.slug, p.image_url, p.status, p.start_date, p.end_date;


-- name: GetPromotionBySlug :one
SELECT
    id,
    title,
    description,
    image_url,
    created_at,
    updated_at,
    start_date,
    end_date,
    slug,
    status
FROM
    promotions
WHERE
    slug = $1
  AND status = 'active'
  AND start_date <= NOW()
  AND end_date >= NOW()
ORDER BY
    updated_at DESC
LIMIT 1;

-- name: AddProductToPromotion :exec
INSERT INTO promotion_products (promotion_id, product_id)
VALUES ($1, $2)
ON CONFLICT (promotion_id, product_id) DO NOTHING;

-- name: AddProductsToPromotion :exec
INSERT INTO promotion_products (promotion_id, product_id)
VALUES ($1, unnest($2::uuid[]))
ON CONFLICT (promotion_id, product_id) DO NOTHING;

-- name: UpdatePromotionImage :exec
UPDATE promotions
SET image_url = $2
WHERE id = $1;

-- name: UpdatePromotion :exec
UPDATE promotions
SET title = $2, description = $3, start_date = $4, end_date = $5, status = $6, image_url = $7
WHERE id = $1;

-- name: GetProductIDsByPromotionID :many
SELECT
    product_id
FROM
    promotion_products
WHERE
    promotion_id = $1;

-- name: RemoveProductsFromPromotion :exec
DELETE FROM promotion_products
WHERE promotion_id = $1
  AND product_id = ANY($2::uuid[]);

-- name: GetPromotionDetails :many
-- This function will fetch the promotion details along with associated products.
WITH rate AS (
    SELECT COALESCE(
                   (SELECT rate_to_kes
                    FROM exchange_rates
                    WHERE currency_code = 'USD'
                      AND (valid_to IS NULL OR valid_to >= NOW())
                      AND valid_from <= NOW()
                    ORDER BY valid_from DESC
                    LIMIT 1),
                   135) AS rate_to_kes
)
SELECT
    p.title AS name,
    p.description,
    p.slug,
    p.image_url AS promotion_image_url,
    p.status,
    p.start_date,
    p.end_date,
    pr.slug AS product_slug,
    pr.name AS product_name,
    (pr.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price,
    COALESCE(d.discount_percentage, 0) AS discount,
    COALESCE(pi.image_url, '') AS image_url
FROM
    promotions p
        LEFT JOIN
    promotion_products pp ON pp.promotion_id = p.id
        LEFT JOIN
    products pr ON pr.id = pp.product_id
        LEFT JOIN
    LATERAL (
        SELECT image_url
        FROM product_images
        WHERE product_id = pr.id
        ORDER BY created_at
        LIMIT 1
        ) pi ON true
        LEFT JOIN
    discounts d ON d.product_id = pr.id AND now() BETWEEN d.start_date AND d.end_date
WHERE
    p.slug = $1;


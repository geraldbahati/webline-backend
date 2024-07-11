-- name: CreatePromotion :one
INSERT INTO promotions (title, description, image_url)
VALUES ($1, $2, $3)
RETURNING id, title, description, image_url, created_at, updated_at;

-- name: AddProductToPromotion :exec
INSERT INTO promotion_products (promotion_id, product_id)
VALUES ($1, $2);

-- name: GetPromotionsWithProducts :many
WITH first_images AS (
    SELECT DISTINCT ON (product_id)
        product_id,
        image_url
    FROM
        product_images
    ORDER BY
        product_id,
        created_at
),
     active_discounts AS (
         SELECT
             product_id,
             discount_percentage
         FROM
             discounts
         WHERE
             start_date <= NOW() AND end_date >= NOW()
     )
SELECT
    p.id AS promotion_id,
    p.title,
    p.description,
    p.image_url AS promotion_image_url,
    p.created_at,
    p.updated_at,
    pr.id AS product_id,
    pr.name AS product_name,
    pr.description AS product_description,
    pr.price,
    COALESCE(ad.discount_percentage, 0) AS discount_percentage,
    fi.image_url AS product_image_url
FROM
    promotions p
        JOIN promotion_products pp ON p.id = pp.promotion_id
        JOIN products pr ON pp.product_id = pr.id
        LEFT JOIN first_images fi ON pr.id = fi.product_id
        LEFT JOIN active_discounts ad ON pr.id = ad.product_id
ORDER BY
    p.created_at DESC;



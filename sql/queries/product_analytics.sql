-- name: GetBestSellerProducts :many
SELECT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.status,
    p.featured,
    p.slug,
    SUM(oi.quantity) AS total_sold
FROM
    products p
        JOIN order_items oi ON p.id = oi.product_id
WHERE
    p.status = 'active'
GROUP BY
    p.id
ORDER BY
    total_sold DESC
LIMIT $1;

-- name: GetFeaturedProducts :many
SELECT
    id,
    name,
    description,
    price,
    stock,
    category_id,
    created_at,
    updated_at,
    status,
    featured,
    slug
FROM
    products
WHERE
    featured = true
ORDER BY
    created_at DESC
LIMIT $1;

-- name: GetNewArrivalProducts :many
SELECT
    id,
    name,
    description,
    price,
    stock,
    category_id,
    created_at,
    updated_at,
    status,
    featured,
    slug
FROM
    products
WHERE
    status = 'active'
  AND created_at >= NOW() - INTERVAL '100 days'
ORDER BY
    created_at DESC
LIMIT $1;


-- name: GetDailyDeals :many
WITH first_images AS (
    SELECT DISTINCT ON (product_id) product_id, image_url
    FROM product_images
    ORDER BY product_id, created_at
)
SELECT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug,
    d.discount_percentage,
    d.start_date,
    d.end_date,
    fi.image_url
FROM
    products p
        JOIN discounts d ON p.id = d.product_id
        LEFT JOIN first_images fi ON p.id = fi.product_id
WHERE
    p.status = 'active'
  AND d.start_date <= now()
  AND d.end_date >= now()
ORDER BY
    d.discount_percentage DESC,
    p.created_at DESC;

-- name: CreateProduct :one
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
INSERT INTO products (
    id, name, description, stock, category_id, created_by, updated_by,
    part_number, meta_title, meta_description, meta_keywords,
    status, usd_price, price_per_unit
)
VALUES (
           gen_random_uuid(),
           $1, $2, $3, $4, $5, $6,
           $7, $8, $9, $10,
           $11,
           $12 / (SELECT rate_to_kes FROM rate),
              $13 / (SELECT rate_to_kes FROM rate)
       )
RETURNING *;



-- name: GetProductByID :one
WITH rate AS (
    SELECT COALESCE(
        (SELECT rate_to_kes
         FROM exchange_rates
         WHERE currency_code = 'USD'
           AND (valid_to IS NULL OR valid_to >= NOW())
           AND valid_from <= NOW()
         ORDER BY valid_from DESC
         LIMIT 1),
        135
    ) AS rate_to_kes
)
SELECT
    p.id,
    p.name,
    p.description,
    p.usd_price,
    (p.usd_price * r.rate_to_kes)::numeric AS price_in_kes,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.search_keyword,
    p.slug
FROM
    products p,
    rate r
WHERE
    p.id = $1;

-- name: GetProductByIDs :many
SELECT id, name, description, usd_price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword, slug
FROM products
WHERE id = ANY($1::uuid[]);

-- name: GetProductBySlug :one
SELECT p.id, p.name, p.description, p.usd_price, p.stock, c.name AS category_name, p.created_at, p.updated_at, p.status, p.created_by, p.updated_by, p.featured, p.search_keyword, p.slug
FROM products p
         JOIN categories c ON p.category_id = c.id
WHERE p.slug = $1;


-- name: ListProducts :many
SELECT id, name, description, usd_price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword, slug
FROM products
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :exec
WITH rate AS (
    SELECT COALESCE(
                   (SELECT er.rate_to_kes
                    FROM exchange_rates er
                    WHERE er.currency_code = 'USD'
                      AND (er.valid_to IS NULL OR er.valid_to >= NOW())
                      AND er.valid_from <= NOW()
                    ORDER BY er.valid_from DESC
                    LIMIT 1),
                   135) AS rate_to_kes
)
UPDATE products p
SET
    name = $1,
    description = $2,
    stock = $3,
    category_id = $4,
    updated_by = $5,
    part_number = $6,
    meta_title = $7,
    meta_description = $8,
    meta_keywords = $9,
    status = $10,
    usd_price = $11 / (SELECT rate_to_kes FROM rate),
    price_per_unit = $12 / (SELECT rate_to_kes FROM rate),
    updated_at = NOW()
WHERE p.id = $13;

-- name: SoftDeleteProduct :exec
UPDATE products
SET status = 'archived', updated_at = NOW()
WHERE id = $1;

-- name: DeleteProductImagesByProductID :exec
DELETE FROM product_images WHERE product_id = $1;

-- name: DeleteProductSpecificationsByProductID :exec
DELETE FROM product_specifications WHERE product_id = $1;

-- name: DeleteProductByID :exec
DELETE FROM products WHERE id = $1;

-- name: ArchiveProductByID :exec
UPDATE products
SET status = 'archived', updated_at = NOW()
WHERE id = $1;


-- name: GetProductsByCategoryID :many
SELECT id, name, description, usd_price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword
FROM products
WHERE category_id = $1
ORDER BY name;

-- name: SearchProducts :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        id,
        name,
        parent_id
    FROM
        categories
    WHERE
        name ILIKE '%' || $1 || '%'
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            INNER JOIN category_hierarchy ch ON c.parent_id = ch.id
)
SELECT DISTINCT ON (p.id)
    p.id,
    p.name,
    p.description,
    p.usd_price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.search_keyword,
    p.slug
FROM
    products p
        LEFT JOIN categories c ON p.category_id = c.id
        LEFT JOIN category_hierarchy ch ON p.category_id = ch.id
WHERE
    (
        p.name ILIKE '%' || $1 || '%' OR
        p.description ILIKE '%' || $1 || '%' OR
        COALESCE(c.name, '') ILIKE '%' || $1 || '%' OR
        p.search_keyword @@ plainto_tsquery('english', $1) OR
        ch.id IS NOT NULL
        )
  AND p.status = 'active'
ORDER BY
    p.id,
    ts_rank(p.search_keyword, plainto_tsquery('english', $1)) DESC,
    p.created_at DESC;


-- name: CountProducts :one
SELECT COUNT(*) AS count
FROM products WHERE status = 'active';

-- name: GetProductsByParentCategoryID :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT
        c2.id,
        c2.name,
        c2.parent_id,
        c2.position
    FROM categories c2
             INNER JOIN category_hierarchy ch ON c2.parent_id = ch.id
)
SELECT
    p.id, p.name, p.description, p.usd_price, p.stock, p.created_at, p.updated_at, p.status,
    p.created_by, p.updated_by, p.featured, p.search_keyword, p.slug,
    c.name AS category_name
FROM products p
         JOIN categories c ON p.category_id = c.id
WHERE p.category_id IN (SELECT ch.id FROM category_hierarchy ch)
AND p.status = 'active'
LIMIT $2 OFFSET $3;


-- name: CountProductsByParentCategoryID :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT
        c2.id,
        c2.name,
        c2.parent_id,
        c2.position
    FROM categories c2
             INNER JOIN category_hierarchy ch ON c2.parent_id = ch.id
)
SELECT COUNT(*) AS count
FROM products p
WHERE p.category_id IN (SELECT ch.id FROM category_hierarchy ch)
AND p.status = 'active';

-- name: GetV2Products :many
WITH first_image AS (
    SELECT DISTINCT ON (product_id) product_id, image_url
    FROM product_images
    ORDER BY product_id, position
),
     rate AS (
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
    p.name,
    (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
    p.status,
    COALESCE(fi.image_url, '') AS imageURL,
    COALESCE(d.discount_percentage, 0) AS discount,
    p.slug,
    p.created_at AS createdAt,
    EXISTS (
        SELECT 1
        FROM promotion_products pp
        WHERE pp.product_id = p.id
    ) AS inPromotion,
    (
        SELECT COALESCE(SUM(oi.quantity), 0)
        FROM order_items oi
                 JOIN orders o ON oi.order_id = o.id
        WHERE oi.product_id = p.id
          AND o.status = 'pending'
    )::int AS totalSales,
    p.part_number AS partNumber,
    pc.name AS categoryName
FROM products p
         LEFT JOIN first_image fi ON fi.product_id = p.id
         LEFT JOIN discounts d ON d.product_id = p.id AND d.start_date <= NOW() AND d.end_date >= NOW()
         LEFT JOIN categories c ON p.category_id = c.id
         LEFT JOIN categories pc ON c.parent_id = pc.id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountV2Products :one
SELECT COUNT(*) AS count
FROM products;

-- name: GetV2ProductDetailBySlug :one
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
),
     product_cte AS (
         SELECT
             p.id,
             p.name,
             p.description,
             (p.usd_price * r.rate_to_kes)::numeric AS price_in_kes,
             (p.price_per_unit * r.rate_to_kes)::numeric AS price_per_unit_in_kes,
             r.rate_to_kes AS exchange_rate,
             p.slug,
             p.stock,
             p.part_number,
             p.category_id,
             c.parent_id AS parent_category_id,
             p.meta_description,
             p.meta_keywords,
             p.meta_title,
             p.status,
             (NOW() BETWEEN p.valid_from AND COALESCE(p.valid_to, 'infinity'::timestamp)) AS is_valid
         FROM
             products p
                 JOIN categories c ON p.category_id = c.id,
             rate r
         WHERE
             p.slug = $1
     ),
     specs_cte AS (
         SELECT
             ps.product_id,
             json_agg(json_build_object('name', ps.spec_name, 'value', ps.spec_value)) AS specifications
         FROM
             product_specifications ps
                 JOIN product_cte p ON ps.product_id = p.id
         GROUP BY
             ps.product_id
     ),
     images_cte AS (
         SELECT
             pi.product_id,
             json_agg(json_build_object('url', pi.image_url) ORDER BY pi.position) AS images
         FROM
             product_images pi
                 JOIN product_cte p ON pi.product_id = p.id
         GROUP BY
             pi.product_id
     ),
     metafields_cte AS (
         SELECT
             ptav.product_id,
             json_agg(json_build_object('name', pa.name, 'value', pav.value)) AS metafields
         FROM
             product_to_attribute_values ptav
                 JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
                 JOIN product_attributes pa ON pav.attribute_id = pa.id
                 JOIN product_cte p ON ptav.product_id = p.id
         GROUP BY
             ptav.product_id
     )
SELECT
    p.name,
    p.description,
    p.price_in_kes::text AS price,
    p.price_per_unit_in_kes::text AS price_per_unit,
    p.exchange_rate::numeric AS exchange_rate,
    CAST(p.stock AS INTEGER) AS stock,
    p.part_number,
    p.slug,
    p.meta_title,
    p.meta_description,
    p.meta_keywords,
    p.category_id,
    p.parent_category_id,
    p.status,
    p.is_valid::boolean AS is_valid,
    COALESCE(s.specifications, '[]'::json) AS specifications,
    COALESCE(i.images, '[]'::json) AS images,
    COALESCE(m.metafields, '[]'::json) AS product_metafields
FROM
    product_cte p
        LEFT JOIN
    specs_cte s ON p.id = s.product_id
        LEFT JOIN
    images_cte i ON p.id = i.product_id
        LEFT JOIN
    metafields_cte m ON p.id = m.product_id;

-- name: ArchiveProductsBySlugs :exec
UPDATE products
SET status = 'archived', updated_by = $2
WHERE slug = ANY($1::text[]);

-- name: ActivateProductsBySlugs :exec
UPDATE products
SET status = 'active', updated_by = $2
WHERE slug = ANY($1::text[]);

-- name: DraftProductsBySlugs :exec
UPDATE products
SET status = 'draft', updated_by = $2
WHERE slug = ANY($1::text[]);

-- name: DeleteProductsBySlugs :exec
DELETE FROM products
WHERE slug = ANY($1::text[]);

-- name: GetProductPricingByProductID :one
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
    p.id,
    p.name,
    p.description,
    p.slug,
    (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
    COALESCE(d.discount_percentage, 0) AS discount_percent,
    COALESCE(pi.image_url, '')::TEXT AS imageUrl
FROM
    products p
        LEFT JOIN LATERAL (
        SELECT
            pi.image_url
        FROM
            product_images pi
        WHERE
            pi.product_id = p.id
        ORDER BY
            pi.position ASC
        LIMIT 1
        ) pi ON TRUE
        LEFT JOIN discounts d ON d.product_id = p.id
WHERE
    p.id = $1
LIMIT 1;

-- name: GetProductSpecsByID :one
SELECT
    p.description,
    json_agg(
            json_build_object(
                    'name', ps.spec_name,
                    'value', ps.spec_value
            )
    ) AS specs
FROM
    products p
        JOIN
    product_specifications ps ON ps.product_id = p.id
WHERE
    p.id = $1
GROUP BY
    p.description;

-- name: GetProductIDsBySlugs :many
SELECT
    id
FROM
    products
WHERE
    slug = ANY($1::text[]);

-- name: GetProductSlugByProductID :one
SELECT
    slug
FROM
    products
WHERE
    id = $1;

-- name: GetProductCartByProductSlug :one
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
    p.id,
    p.name,
    p.description,
    (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price,
    p.stock,
    p.category_id,
    p.featured,
    COALESCE(d.discount_percentage, 0) AS discount_percent,
    p.slug
FROM
    products p
        LEFT JOIN discounts d ON d.product_id = p.id
WHERE
    p.slug = $1;

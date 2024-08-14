-- -- name: CountFilteredCategoryProducts :one
-- -- $1: selected_category_names (VARCHAR[]) sqlc.arg(selectedCategoryNames)
-- -- $2: min_price_kes (NUMERIC) sqlc.arg(minPriceKES)
-- -- $3: max_price_kes (NUMERIC) sqlc.arg(maxPriceKES)
-- -- $4: selected_attribute_names (VARCHAR[]) sqlc.arg(selectedAttributeNames)
-- -- $5: selected_attribute_values (VARCHAR[]) sqlc.arg(selectedAttributeValues)
-- WITH RECURSIVE category_hierarchy AS (
--     SELECT
--         c.id,
--         c.name,
--         c.parent_id
--     FROM
--         categories c
--     WHERE
--         c.name = ANY($1::VARCHAR[]) -- Start with the given list of category names
--
--     UNION ALL
--
--     SELECT
--         c.id,
--         c.name,
--         c.parent_id
--     FROM
--         categories c
--             INNER JOIN
--         category_hierarchy ch ON c.parent_id = ch.id
-- ),
--                current_exchange_rate AS (
--                    SELECT
--                        COALESCE((
--                                     SELECT rate_to_kes
--                                     FROM exchange_rates
--                                     WHERE currency_code = 'USD'
--                                       AND (valid_to IS NULL OR valid_to >= NOW())
--                                     ORDER BY valid_from DESC
--                                     LIMIT 1
--                                 ), 135) AS rate_to_kes
--                ),
--                filtered_products AS (
--                    SELECT
--                        p.id
--                    FROM
--                        products p
--                            LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
--                            LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
--                            LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
--                        current_exchange_rate er
--                    WHERE
--                        p.category_id IN (SELECT id FROM category_hierarchy)
--                      -- Attribute filters
--                      AND (
--                        (
--                            pa.name = ANY($4::VARCHAR[]) AND
--                            pav.value = ANY($5::VARCHAR[])
--                            )
--                            OR $4 IS NULL
--                        )
--                      -- Price range filter (converted to KES)
--                      AND (p.usd_price * er.rate_to_kes) BETWEEN $2::numeric AND $3::numeric
--                      -- Status filter
--                      AND p.status = 'active'
--                )
-- SELECT COUNT(*) AS total_count FROM filtered_products;


-- name: GetTotalCategoryProductsByFilters :one
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT COUNT(DISTINCT p.id) AS total_count
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
    jsonb_typeof($5::jsonb) IS NULL
        OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL
        AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
    )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active';

-- name: GetCategoryProductsByFiltersPriceAsc :many
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $6: limit (INTEGER)
-- $7: offset (INTEGER)
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    (p.usd_price * rate.rate_to_kes)::numeric AS price,
    p.slug,
    COALESCE(
            (SELECT MAX(d.discount_percentage)
             FROM discounts d
             WHERE d.product_id = p.id
               AND d.start_date <= NOW()
               AND (d.end_date IS NULL OR d.end_date >= NOW())
            ), 0)::numeric AS discountPercent,
    (SELECT pi.image_url
     FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.position ASC
     LIMIT 1) AS imageURL
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
      jsonb_typeof($5::jsonb) IS NULL 
      OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL 
      AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
  )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active'
ORDER BY p.usd_price ASC
LIMIT $6 OFFSET $7;

-- name: GetCategoryProductsByFiltersPriceDesc :many
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $6: limit (INTEGER)
-- $7: offset (INTEGER)
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    (p.usd_price * rate.rate_to_kes)::numeric AS price,
    p.slug,
    COALESCE(
            (SELECT MAX(d.discount_percentage)
             FROM discounts d
             WHERE d.product_id = p.id
               AND d.start_date <= NOW()
               AND (d.end_date IS NULL OR d.end_date >= NOW())
            ), 0)::numeric AS discountPercent,
    (SELECT pi.image_url
     FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.position ASC
     LIMIT 1) AS imageURL
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
      jsonb_typeof($5::jsonb) IS NULL 
      OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL 
      AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
  )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active'
ORDER BY p.usd_price DESC
LIMIT $6 OFFSET $7;

-- name: GetCategoryProductsByFiltersNameAsc :many
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $6: limit (INTEGER)
-- $7: offset (INTEGER)
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    (p.usd_price * rate.rate_to_kes)::numeric AS price,
    p.slug,
    COALESCE(
            (SELECT MAX(d.discount_percentage)
             FROM discounts d
             WHERE d.product_id = p.id
               AND d.start_date <= NOW()
               AND (d.end_date IS NULL OR d.end_date >= NOW())
            ), 0)::numeric AS discountPercent,
    (SELECT pi.image_url
     FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.position ASC
     LIMIT 1) AS imageURL
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
      jsonb_typeof($5::jsonb) IS NULL 
      OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL 
      AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
  )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active'
ORDER BY p.name ASC
LIMIT $6 OFFSET $7;

-- name: GetCategoryProductsByFiltersNameDesc :many
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $6: limit (INTEGER)
-- $7: offset (INTEGER)
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    (p.usd_price * rate.rate_to_kes)::numeric AS price,
    p.slug,
    COALESCE(
            (SELECT MAX(d.discount_percentage)
             FROM discounts d
             WHERE d.product_id = p.id
               AND d.start_date <= NOW()
               AND (d.end_date IS NULL OR d.end_date >= NOW())
            ), 0)::numeric AS discountPercent,
    (SELECT pi.image_url
     FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.position ASC
     LIMIT 1) AS imageURL
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
      jsonb_typeof($5::jsonb) IS NULL 
      OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL 
      AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
  )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active'
ORDER BY p.name DESC
LIMIT $6 OFFSET $7;

-- name: GetCategoryProductsByFiltersNewest :many
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $6: limit (INTEGER)
-- $7: offset (INTEGER)
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    (p.usd_price * rate.rate_to_kes)::numeric AS price,
    p.slug,
    COALESCE(
            (SELECT MAX(d.discount_percentage)
             FROM discounts d
             WHERE d.product_id = p.id
               AND d.start_date <= NOW()
               AND (d.end_date IS NULL OR d.end_date >= NOW())
            ), 0)::numeric AS discountPercent,
    (SELECT pi.image_url
     FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.position ASC
     LIMIT 1) AS imageURL
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
      jsonb_typeof($5::jsonb) IS NULL 
      OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL 
      AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
  )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active'
ORDER BY p.created_at DESC
LIMIT $6 OFFSET $7;

-- name: GetCategoryProductsByFiltersOldest :many
-- $1: category_id (UUID)
-- $2: selected_category_names (VARCHAR[])
-- $3: min_price_kes (NUMERIC)
-- $4: max_price_kes (NUMERIC)
-- $5: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $6: limit (INTEGER)
-- $7: offset (INTEGER)
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
), rate AS (
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
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    (p.usd_price * rate.rate_to_kes)::numeric AS price,
    p.slug,
    COALESCE(
            (SELECT MAX(d.discount_percentage)
             FROM discounts d
             WHERE d.product_id = p.id
               AND d.start_date <= NOW()
               AND (d.end_date IS NULL OR d.end_date >= NOW())
            ), 0)::numeric AS discountPercent,
    (SELECT pi.image_url
     FROM product_images pi
     WHERE pi.product_id = p.id
     ORDER BY pi.position ASC
     LIMIT 1) AS imageURL
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
         LEFT JOIN product_to_attribute_values ptav ON p.id = ptav.product_id
         LEFT JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
         LEFT JOIN product_attributes pa ON pav.attribute_id = pa.id,
     rate
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (
      jsonb_typeof($5::jsonb) IS NULL 
      OR jsonb_each_text($5::jsonb) ->> pa.name IS NOT NULL 
      AND pav.value = (jsonb_each_text($5::jsonb) ->> pa.name)
  )
  AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $3 AND $4
  AND p.status = 'active'
ORDER BY p.created_at ASC
LIMIT $6 OFFSET $7;

-- name: GetProductAttributesByCategoryID :one
WITH category_products AS (
    SELECT p.id
    FROM products p
    WHERE p.category_id = $1 AND p.status = 'active'
),
     attribute_values AS (
         SELECT
             pa.name AS attribute_name,
             pav.value AS attribute_value
         FROM
             product_to_attribute_values ptav
                 JOIN
             product_attribute_values pav ON ptav.attribute_value_id = pav.id
                 JOIN
             product_attributes pa ON pav.attribute_id = pa.id
         WHERE
             ptav.product_id IN (SELECT id FROM category_products)
     )
SELECT
    jsonb_agg(
            jsonb_build_object(
                    attribute_name, jsonb_agg(DISTINCT attribute_value)
            )
    ) AS attributes,
    (SELECT COUNT(*) FROM category_products) AS total_products
FROM
    attribute_values
GROUP BY
    attribute_name;



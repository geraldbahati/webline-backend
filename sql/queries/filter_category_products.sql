
-- name: GetTotalCategoryProductsByFilters :one
-- $1: category_id (UUID)
-- $2: min_price_kes (NUMERIC)
-- $3: max_price_kes (NUMERIC)
-- $4: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}

WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    COUNT(fp.id) AS total_products
FROM
    filtered_products fp;

-- name: GetCategoryProductsByFiltersPriceAsc :many
-- $1: category_id (UUID)
-- $2: min_price_kes (NUMERIC)
-- $3: max_price_kes (NUMERIC)
-- $4: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $5: limit (INTEGER)
-- $6: offset (INTEGER)
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
                       p.created_at,
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug,
    fp.created_at
FROM
    filtered_products fp
ORDER BY
    fp.price_in_kes ASC
LIMIT
    $5 OFFSET $6;

-- name: GetCategoryProductsByFiltersPriceDesc :many
-- $1: category_id (UUID)
-- $2: min_price_kes (NUMERIC)
-- $3: max_price_kes (NUMERIC)
-- $4: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $5: limit (INTEGER)
-- $6: offset (INTEGER)
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
                       p.created_at,
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug,
    fp.created_at
FROM
    filtered_products fp
ORDER BY
    fp.price_in_kes DESC
LIMIT
    $5 OFFSET $6;

-- name: GetCategoryProductsByFiltersNameAsc :many
-- $1: category_id (UUID)
-- $2: min_price_kes (NUMERIC)
-- $3: max_price_kes (NUMERIC)
-- $4: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $5: limit (INTEGER)
-- $6: offset (INTEGER)
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
                       p.created_at,
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug,
    fp.created_at
FROM
    filtered_products fp
ORDER BY
    fp.name ASC
LIMIT
    $5 OFFSET $6;

-- name: GetCategoryProductsByFiltersNameDesc :many
-- $1: category_id (UUID)
-- $2: min_price_kes (NUMERIC)
-- $3: max_price_kes (NUMERIC)
-- $4: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $5: limit (INTEGER)
-- $6: offset (INTEGER)
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
                       p.created_at,
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug,
    fp.created_at
FROM
    filtered_products fp
ORDER BY
    fp.name DESC
LIMIT
    $5 OFFSET $6;

-- name: GetCategoryProductsByFiltersNewest :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
                       p.created_at,
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug,
    fp.created_at
FROM
    filtered_products fp
ORDER BY
    fp.created_at DESC
LIMIT
    $5 OFFSET $6;

-- name: GetCategoryProductsByFiltersOldest :many
-- $1: category_id (UUID)
-- $2: min_price_kes (NUMERIC)
-- $3: max_price_kes (NUMERIC)
-- $4: selected_attributes (JSONB) -- {"size": ["14-inch", "15-inch"], "color": ["red", "blue"]}
-- $5: limit (INTEGER)
-- $6: offset (INTEGER)
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.id = $1
      AND c.is_active = true
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
    WHERE c.is_active = true
),
               filtered_categories AS (
                   SELECT
                       c.id
                   FROM
                       categories c
                   WHERE
                       c.id IN (SELECT id FROM category_hierarchy)
                     AND (
                       ($4::jsonb IS NULL OR $4 = '{}'::jsonb)
                           OR LOWER(TRIM(c.name)) IN (
                           SELECT LOWER(TRIM(value::text))
                           FROM jsonb_array_elements_text($4::jsonb->'Brand') AS value
                       )
                       )
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
               ),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric AS price_in_kes,
                       p.created_at,
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       COALESCE(
                               (SELECT pi.image_url
                                FROM product_images pi
                                WHERE pi.product_id = p.id
                                ORDER BY pi.position ASC
                                LIMIT 1), '')::text AS image_url
                   FROM
                       products p
                   WHERE
                       p.category_id IN (SELECT id FROM filtered_categories)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug,
    fp.created_at
FROM
    filtered_products fp
ORDER BY
    fp.price_in_kes DESC
LIMIT
    $5 OFFSET $6;

-- name: GetProductAttributesByCategoryID :one
WITH RECURSIVE category_tree AS (
    -- Recursive CTE to get all descendants of the parent category
    SELECT c.id, c.name, c.parent_id, c.position
    FROM categories c
    WHERE c.id = $1  -- Parent category ID
    UNION
    SELECT c.id, c.name, c.parent_id, c.position
    FROM categories c
             INNER JOIN category_tree ct ON c.parent_id = ct.id
),
               category_products AS (
                   -- Get all products that belong to the descendant categories
                   SELECT p.id
                   FROM products p
                   WHERE p.category_id IN (SELECT ct.id FROM category_tree ct) AND p.status = 'active'
               ),
               attribute_values AS (
                   -- Get the product attribute values
                   SELECT
                       pa.name AS attribute_name,
                       pav.value AS attribute_value,
                       c.position AS attribute_position
                   FROM
                       product_to_attribute_values ptav
                           JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN product_attributes pa ON pav.attribute_id = pa.id
                           LEFT JOIN categories c ON c.name = pav.value  -- Assuming attribute values are tied to categories for position
                   WHERE
                       ptav.product_id IN (SELECT cp.id FROM category_products cp)
                   UNION ALL
                   -- Include brand as an attribute
                   SELECT
                       'Brand' AS attribute_name,
                       c.name AS attribute_value,
                       c.position AS attribute_position
                   FROM
                       category_tree ct
                           JOIN categories c ON c.parent_id = ct.id
               ),
               aggregated_attributes AS (
                   -- Aggregate attribute values
                   SELECT
                       av.attribute_name,
                       jsonb_agg(av.attribute_value ORDER BY av.attribute_position, av.attribute_value) AS aggregated_values
                   FROM
                       attribute_values av
                   GROUP BY
                       av.attribute_name
               )
SELECT
    jsonb_object_agg(aa.attribute_name, aa.aggregated_values) AS attributes,
    (SELECT COUNT(*) FROM category_products cp) AS total_products
FROM (
         SELECT
             attribute_name,
             aggregated_values
         FROM
             aggregated_attributes
         UNION ALL
         -- Ensure the "Brand" attribute is included without depending on the join with categories
         SELECT
             'Brand' AS attribute_name,
             (SELECT aggregated_values FROM aggregated_attributes WHERE attribute_name = 'Brand') AS aggregated_values
     ) AS aa;


-- name: GetProductAttributesByCategoryName :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = $1

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
),
               category_products AS (
                   SELECT p.id
                   FROM products p
                   WHERE p.category_id IN (SELECT id FROM category_hierarchy) AND p.status = 'active'
               ),
               brand_attributes AS (
                   SELECT
                       'brand' AS attribute_name,
                       c.name AS attribute_value,
                       c.position
                   FROM
                       categories c
                   WHERE
                       c.parent_id = (SELECT id FROM category_hierarchy LIMIT 1) -- Get the immediate children of the main category
               ),
               attribute_values AS (
                   SELECT
                       pa.name AS attribute_name,
                       pav.value AS attribute_value
                   FROM
                       product_to_attribute_values ptav
                           JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM category_products)
                   UNION ALL
                   SELECT
                       ba.attribute_name,
                       ba.attribute_value
                   FROM
                       brand_attributes ba
               ),
               ordered_attributes AS (
                   SELECT
                       av.attribute_name,
                       av.attribute_value,
                       ba.position
                   FROM
                       attribute_values av
                           LEFT JOIN
                       brand_attributes ba ON av.attribute_value = ba.attribute_value AND av.attribute_name = 'brand'
                   ORDER BY
                       av.attribute_name, ba.position
               ),
               aggregated_attributes AS (
                   SELECT
                       attribute_name,
                       jsonb_agg(attribute_value) AS aggregated_values
                   FROM
                       ordered_attributes
                   GROUP BY
                       attribute_name
               )
SELECT
    jsonb_object_agg(
            attribute_name, aggregated_values
    ) AS attributes,
    (SELECT COUNT(*) FROM category_products) AS total_products
FROM
    aggregated_attributes;








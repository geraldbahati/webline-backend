-- name: GetAllProductsByFiltersPriceAsc :many
WITH category_ids AS (
    -- Use materialized path pattern instead of recursive CTE
    SELECT c.id
    FROM categories c
    WHERE
        -- Find categories that match the filter names directly
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR c.name = ANY (
            COALESCE(ARRAY(SELECT jsonb_object_keys($3::jsonb)), ARRAY[]::text[])
        )
    UNION
    -- Find all descendant categories using ltree path
    SELECT c.id
    FROM categories root
    JOIN categories c ON c.path <@ root.path
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR root.name = ANY (
            COALESCE(ARRAY(SELECT jsonb_object_keys($3::jsonb)), ARRAY[]::text[])
        )
),
rate AS (
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
),
filtered_products AS (
    SELECT
        p.id,
        p.name,
        p.description,
        (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
        p.created_at,
        p.slug,
        COALESCE(d.discount_percent, 0)::numeric AS discount_percent,
        COALESCE(img.image_url, '')::text AS image_url
    FROM products p
    CROSS JOIN rate
    /*-- LATERAL join to efficiently compute the maximum discount; executes once per product row --*/
    LEFT JOIN LATERAL (
        SELECT MAX(d.discount_percentage) AS discount_percent
        FROM discounts d
        WHERE d.product_id = p.id
          AND d.start_date <= NOW()
          AND (d.end_date IS NULL OR d.end_date >= NOW())
    ) d ON true
    /*-- LATERAL join to fetch the first image URL per product --*/
    LEFT JOIN LATERAL (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
    ) img ON true
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
has_attributes AS (
    SELECT
        EXISTS (SELECT 1 FROM product_attributes LIMIT 1) AS has_product_attributes,
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb) AS is_empty_filter
),
final_products AS (
    SELECT fp.*
    FROM filtered_products fp, has_attributes ha
    WHERE
        ha.is_empty_filter
        OR NOT ha.has_product_attributes
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
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
FROM final_products fp
ORDER BY fp.price_in_kes ASC
LIMIT $4 OFFSET $5;



-- name: GetAllProductsByFiltersPriceDesc :many
WITH RECURSIVE category_ids AS (
    SELECT id
    FROM categories
    WHERE ($3::jsonb IS NULL OR $3 = '{}'::jsonb OR name = ANY(SELECT jsonb_object_keys($3::jsonb)))
    UNION ALL
    SELECT c.id
    FROM categories c
    JOIN category_ids ci ON c.parent_id = ci.id
),
rate AS (
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
),
filtered_products AS (
    SELECT
        p.id,
        p.name,
        p.description,
        (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
        p.created_at,
        p.slug,
        COALESCE(
            (
              SELECT MAX(d.discount_percentage)
              FROM discounts d
              WHERE d.product_id = p.id
                AND d.start_date <= NOW()
                AND (d.end_date IS NULL OR d.end_date >= NOW())
            ),
            0
        )::numeric AS discount_percent,
        COALESCE(
            (
              SELECT pi.image_url
              FROM product_images pi
              WHERE pi.product_id = p.id
              ORDER BY pi.position ASC
              LIMIT 1
            ),
            ''
        )::text AS image_url
    FROM products p
    CROSS JOIN rate
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
final_products AS (
    SELECT fp.*
    FROM filtered_products fp
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
        OR NOT EXISTS (SELECT 1 FROM product_attributes)
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
FROM final_products fp
ORDER BY fp.price_in_kes DESC
LIMIT $4 OFFSET $5;


-- name: GetAllProductsByFiltersNameAsc :many
WITH RECURSIVE category_ids AS (
    SELECT id
    FROM categories
    WHERE ($3::jsonb IS NULL OR $3 = '{}'::jsonb OR name = ANY(SELECT jsonb_object_keys($3::jsonb)))
    UNION ALL
    SELECT c.id
    FROM categories c
    JOIN category_ids ci ON c.parent_id = ci.id
),
rate AS (
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
),
filtered_products AS (
    SELECT
        p.id,
        p.name,
        p.description,
        (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
        p.created_at,
        p.slug,
        COALESCE(
            (
              SELECT MAX(d.discount_percentage)
              FROM discounts d
              WHERE d.product_id = p.id
                AND d.start_date <= NOW()
                AND (d.end_date IS NULL OR d.end_date >= NOW())
            ),
            0
        )::numeric AS discount_percent,
        COALESCE(
            (
              SELECT pi.image_url
              FROM product_images pi
              WHERE pi.product_id = p.id
              ORDER BY pi.position ASC
              LIMIT 1
            ),
            ''
        )::text AS image_url
    FROM products p
    CROSS JOIN rate
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
final_products AS (
    SELECT fp.*
    FROM filtered_products fp
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
        OR NOT EXISTS (SELECT 1 FROM product_attributes)
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
FROM final_products fp
ORDER BY fp.name ASC
LIMIT $4 OFFSET $5;


-- name: GetAllProductsByFiltersNameDesc :many
WITH RECURSIVE category_ids AS (
    SELECT id
    FROM categories
    WHERE ($3::jsonb IS NULL OR $3 = '{}'::jsonb OR name = ANY(SELECT jsonb_object_keys($3::jsonb)))
    UNION ALL
    SELECT c.id
    FROM categories c
    JOIN category_ids ci ON c.parent_id = ci.id
),
rate AS (
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
),
filtered_products AS (
    SELECT
        p.id,
        p.name,
        p.description,
        (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
        p.created_at,
        p.slug,
        COALESCE(
            (
              SELECT MAX(d.discount_percentage)
              FROM discounts d
              WHERE d.product_id = p.id
                AND d.start_date <= NOW()
                AND (d.end_date IS NULL OR d.end_date >= NOW())
            ),
            0
        )::numeric AS discount_percent,
        COALESCE(
            (
              SELECT pi.image_url
              FROM product_images pi
              WHERE pi.product_id = p.id
              ORDER BY pi.position ASC
              LIMIT 1
            ),
            ''
        )::text AS image_url
    FROM products p
    CROSS JOIN rate
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
final_products AS (
    SELECT fp.*
    FROM filtered_products fp
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
        OR NOT EXISTS (SELECT 1 FROM product_attributes)
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
FROM final_products fp
ORDER BY fp.name DESC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersNewest :many
WITH RECURSIVE category_ids AS (
    SELECT id
    FROM categories
    WHERE ($3::jsonb IS NULL OR $3 = '{}'::jsonb OR name = ANY(SELECT jsonb_object_keys($3::jsonb)))
    UNION ALL
    SELECT c.id
    FROM categories c
    JOIN category_ids ci ON c.parent_id = ci.id
),
rate AS (
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
),
filtered_products AS (
    SELECT
        p.id,
        p.name,
        p.description,
        (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
        p.created_at,
        p.slug,
        COALESCE(
            (
              SELECT MAX(d.discount_percentage)
              FROM discounts d
              WHERE d.product_id = p.id
                AND d.start_date <= NOW()
                AND (d.end_date IS NULL OR d.end_date >= NOW())
            ),
            0
        )::numeric AS discount_percent,
        COALESCE(
            (
              SELECT pi.image_url
              FROM product_images pi
              WHERE pi.product_id = p.id
              ORDER BY pi.position ASC
              LIMIT 1
            ),
            ''
        )::text AS image_url
    FROM products p
    CROSS JOIN rate
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
final_products AS (
    SELECT fp.*
    FROM filtered_products fp
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
        OR NOT EXISTS (SELECT 1 FROM product_attributes)
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
FROM final_products fp
ORDER BY fp.created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersOldest :many
WITH RECURSIVE category_ids AS (
    SELECT id
    FROM categories
    WHERE ($3::jsonb IS NULL OR $3 = '{}'::jsonb OR name = ANY(SELECT jsonb_object_keys($3::jsonb)))
    UNION ALL
    SELECT c.id
    FROM categories c
    JOIN category_ids ci ON c.parent_id = ci.id
),
rate AS (
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
),
filtered_products AS (
    SELECT
        p.id,
        p.name,
        p.description,
        (p.usd_price * rate.rate_to_kes)::numeric AS price_in_kes,
        p.created_at,
        p.slug,
        COALESCE(
            (
              SELECT MAX(d.discount_percentage)
              FROM discounts d
              WHERE d.product_id = p.id
                AND d.start_date <= NOW()
                AND (d.end_date IS NULL OR d.end_date >= NOW())
            ),
            0
        )::numeric AS discount_percent,
        COALESCE(
            (
              SELECT pi.image_url
              FROM product_images pi
              WHERE pi.product_id = p.id
              ORDER BY pi.position ASC
              LIMIT 1
            ),
            ''
        )::text AS image_url
    FROM products p
    CROSS JOIN rate
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
final_products AS (
    SELECT fp.*
    FROM filtered_products fp
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
        OR NOT EXISTS (SELECT 1 FROM product_attributes)
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
FROM final_products fp
ORDER BY fp.created_at ASC
LIMIT $4 OFFSET $5;


-- name: CountAllProductsByFilters :one
WITH category_ids AS (
    -- Use materialized path pattern instead of recursive CTE
    SELECT c.id
    FROM categories c
    WHERE
        -- Find categories that match the filter names directly
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR c.name = ANY (
            COALESCE(ARRAY(SELECT jsonb_object_keys($3::jsonb)), ARRAY[]::text[])
        )
    UNION
    -- Find all descendant categories using ltree path
    SELECT c.id
    FROM categories root
    JOIN categories c ON c.path <@ root.path
    WHERE
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb)
        OR root.name = ANY (
            COALESCE(ARRAY(SELECT jsonb_object_keys($3::jsonb)), ARRAY[]::text[])
        )
),
rate AS (
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
),
-- Filter to only active products in the price range
filtered_products AS (
    SELECT p.id
    FROM products p, rate
    WHERE p.category_id IN (SELECT id FROM category_ids)
      AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
      AND p.status = 'active'
      AND p.valid_from <= NOW()
      AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
-- Optimize attribute filtering - avoid expensive checks when not needed
has_attributes AS (
    SELECT
        EXISTS (SELECT 1 FROM product_attributes LIMIT 1) AS has_product_attributes,
        ($3::jsonb IS NULL OR $3 = '{}'::jsonb) AS is_empty_filter
),
-- Do attribute filtering when needed
final_filtered AS (
    SELECT fp.id
    FROM filtered_products fp, has_attributes ha
    WHERE
        ha.is_empty_filter
        OR NOT ha.has_product_attributes
        OR EXISTS (
            SELECT 1
            FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
            JOIN product_to_attribute_values ptav ON ptav.product_id = fp.id
            JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
            JOIN product_attributes pa ON pav.attribute_id = pa.id
            WHERE pa.name = filter.attr_name
              AND pav.value = filter.attr_value
        )
)
SELECT COUNT(*) AS total_products
FROM final_filtered;



-- name: GetProductAttributes :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM categories c
    WHERE c.parent_id IS NULL   -- Root categories
    UNION ALL
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM categories c
    JOIN category_hierarchy ch ON c.parent_id = ch.id
),
active_products AS (
    SELECT id
    FROM products
    WHERE status = 'active'
      AND valid_from <= NOW()
      AND (valid_to IS NULL OR valid_to >= NOW())
),
attribute_values AS (
    SELECT
        pa.name AS attribute_name,
        pav.value AS attribute_value,
        COALESCE(c.position, 999) AS attribute_position
    FROM product_to_attribute_values ptav
    JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
    JOIN product_attributes pa ON pav.attribute_id = pa.id
    LEFT JOIN categories c ON c.name = pav.value
    WHERE EXISTS (
        SELECT 1
        FROM active_products
        WHERE id = ptav.product_id
    )
    UNION ALL
    SELECT
        ch.name AS attribute_name,
        c.name AS attribute_value,
        c.position AS attribute_position
    FROM category_hierarchy ch
    JOIN categories c ON c.parent_id = ch.id
    WHERE ch.parent_id IS NULL
),
aggregated_attributes AS (
    SELECT
        attribute_name,
        jsonb_agg(attribute_value ORDER BY attribute_position, attribute_value) AS aggregated_values
    FROM attribute_values
    GROUP BY attribute_name
)
SELECT
    jsonb_object_agg(attribute_name, aggregated_values) AS attributes,
    (SELECT COUNT(*) FROM active_products) AS total_products
FROM aggregated_attributes;

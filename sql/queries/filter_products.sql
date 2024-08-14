-- name: GetAllProductsByFiltersPriceAsc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       (SELECT pi.image_url
                        FROM product_images pi
                        WHERE pi.product_id = p.id
                        ORDER BY pi.position ASC
                        LIMIT 1) AS image_url
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           JOIN
                       filtered_attributes fa ON fp.id = fa.product_id
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug
FROM
    final_products fp
ORDER BY
    fp.price_in_kes ASC
LIMIT
    $5 OFFSET $6;


-- name: GetAllProductsByFiltersPriceDesc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       (SELECT pi.image_url
                        FROM product_images pi
                        WHERE pi.product_id = p.id
                        ORDER BY pi.position ASC
                        LIMIT 1) AS image_url
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           JOIN
                       filtered_attributes fa ON fp.id = fa.product_id
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug
FROM
    final_products fp
ORDER BY
    fp.price_in_kes DESC
LIMIT
    $5 OFFSET $6;

-- name: GetAllProductsByFiltersNameAsc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       (SELECT pi.image_url
                        FROM product_images pi
                        WHERE pi.product_id = p.id
                        ORDER BY pi.position ASC
                        LIMIT 1) AS image_url
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           JOIN
                       filtered_attributes fa ON fp.id = fa.product_id
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug
FROM
    final_products fp
ORDER BY
    fp.name ASC
LIMIT
    $5 OFFSET $6;

-- name: GetAllProductsByFiltersNameDesc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       (SELECT pi.image_url
                        FROM product_images pi
                        WHERE pi.product_id = p.id
                        ORDER BY pi.position ASC
                        LIMIT 1) AS image_url
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           JOIN
                       filtered_attributes fa ON fp.id = fa.product_id
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price_in_kes AS price,
    fp.image_url AS imageURL,
    fp.discount_percent AS discountPercent,
    fp.slug
FROM
    final_products fp
ORDER BY
    fp.name DESC
LIMIT
    $5 OFFSET $6;


-- name: GetAllProductsByFiltersNewest :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       (SELECT pi.image_url
                        FROM product_images pi
                        WHERE pi.product_id = p.id
                        ORDER BY pi.position ASC
                        LIMIT 1) AS image_url,
                       p.created_at
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           JOIN
                       filtered_attributes fa ON fp.id = fa.product_id
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
    final_products fp
ORDER BY
    fp.created_at DESC
LIMIT
    $5 OFFSET $6;


-- name: GetAllProductsByFiltersOldest :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.slug,
                       COALESCE(
                               (SELECT MAX(d.discount_percentage)
                                FROM discounts d
                                WHERE d.product_id = p.id
                                  AND d.start_date <= NOW()
                                  AND (d.end_date IS NULL OR d.end_date >= NOW())
                               ), 0)::numeric AS discount_percent,
                       (SELECT pi.image_url
                        FROM product_images pi
                        WHERE pi.product_id = p.id
                        ORDER BY pi.position ASC
                        LIMIT 1) AS image_url,
                       p.created_at
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           JOIN
                       filtered_attributes fa ON fp.id = fa.product_id
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
    final_products fp
ORDER BY
    fp.created_at ASC
LIMIT
    $5 OFFSET $6;

-- name: CountAllProductsByFilters :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            JOIN
        category_hierarchy ch ON c.parent_id = ch.id
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
                       p.id
                   FROM
                       products p
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $2 AND $3
                     AND p.status = 'active'
               ),
               filtered_attributes AS (
                   SELECT
                       ptav.product_id
                   FROM
                       product_to_attribute_values ptav
                           JOIN
                       product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN
                       product_attributes pa ON pav.attribute_id = pa.id
                   WHERE
                       ptav.product_id IN (SELECT id FROM filtered_products)
                     AND (
                       jsonb_typeof($4::jsonb) IS NULL
                           OR EXISTS (
                           SELECT 1
                           FROM jsonb_each_text($4::jsonb) AS filter(attr_name, attr_value)
                           WHERE pa.name = filter.attr_name
                             AND pav.value = filter.attr_value
                       )
                       )
                   GROUP BY
                       ptav.product_id
                   HAVING
                       COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($4::jsonb))
               )
SELECT
    COUNT(DISTINCT fa.product_id) AS total_products
FROM
    filtered_attributes fa;

-- name: GetProductAttributes :one
WITH all_active_products AS (
    SELECT p.id
    FROM products p
    WHERE p.status = 'active'
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
             ptav.product_id IN (SELECT id FROM all_active_products)
     )
SELECT
    jsonb_agg(
            jsonb_build_object(
                    attribute_name, jsonb_agg(DISTINCT attribute_value)
            )
    ) AS attributes,
    (SELECT COUNT(*) FROM all_active_products) AS total_products
FROM
    attribute_values
GROUP BY
    attribute_name;

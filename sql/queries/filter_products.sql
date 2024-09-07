-- name: GetAllProductsByFiltersPriceAsc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
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
FROM
    final_products fp
ORDER BY
    fp.price_in_kes ASC
LIMIT
    $4 OFFSET $5;



-- name: GetAllProductsByFiltersPriceDesc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
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
FROM
    final_products fp
ORDER BY
    fp.price_in_kes DESC
LIMIT
    $4 OFFSET $5;


-- name: GetAllProductsByFiltersNameAsc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
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
FROM
    final_products fp
ORDER BY
    fp.name ASC
LIMIT
    $4 OFFSET $5;


-- name: GetAllProductsByFiltersNameDesc :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
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
FROM
    final_products fp
ORDER BY
    fp.name DESC
LIMIT
    $4 OFFSET $5;

-- name: GetAllProductsByFiltersNewest :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
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
FROM
    final_products fp
ORDER BY
    fp.created_at DESC
LIMIT
    $4 OFFSET $5;

-- name: GetAllProductsByFiltersOldest :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.*
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
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
FROM
    final_products fp
ORDER BY
    fp.created_at ASC
LIMIT
    $4 OFFSET $5;


-- name: CountAllProductsByFilters :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        $3::jsonb IS NULL
       OR $3 = '{}'::jsonb
       OR c.name = ANY (
        SELECT jsonb_object_keys($3::jsonb)
    )
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
               sub_category_hierarchy AS (
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                   WHERE
                       $3::jsonb IS NULL
                      OR $3 = '{}'::jsonb
                      OR c.name IN (
                       SELECT jsonb_array_elements_text($3::jsonb -> (
                           SELECT jsonb_object_keys($3::jsonb) LIMIT 1
                       ))
                   )
                       AND c.parent_id IN (SELECT id FROM category_hierarchy)
                   UNION ALL
                   SELECT
                       c.id,
                       c.name,
                       c.parent_id
                   FROM
                       categories c
                           JOIN
                       sub_category_hierarchy sch ON c.parent_id = sch.id
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
                       p.category_id IN (SELECT id FROM sub_category_hierarchy)
                     AND (p.usd_price * (SELECT rate_to_kes FROM rate))::numeric BETWEEN $1 AND $2
                     AND p.status = 'active'
                     AND p.valid_from <= NOW()
                     AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               final_products AS (
                   SELECT DISTINCT
                       fp.id
                   FROM
                       filtered_products fp
                           LEFT JOIN (
                           SELECT
                               ptav.product_id
                           FROM
                               product_to_attribute_values ptav
                                   JOIN
                               product_attribute_values pav ON ptav.attribute_value_id = pav.id
                                   JOIN
                               product_attributes pa ON pav.attribute_id = pa.id
                           WHERE
                               $3::jsonb IS NULL
                              OR $3 = '{}'::jsonb
                              OR EXISTS (
                               SELECT 1
                               FROM jsonb_each_text($3::jsonb) AS filter(attr_name, attr_value)
                               WHERE pa.name = filter.attr_name
                                 AND pav.value = filter.attr_value
                           )
                           GROUP BY
                               ptav.product_id
                           HAVING
                               COUNT(DISTINCT pa.name) = (SELECT COUNT(*) FROM jsonb_each_text($3::jsonb))
                       ) fa ON fp.id = fa.product_id
                   WHERE
                       fa.product_id IS NOT NULL
                      OR NOT EXISTS (
                       SELECT 1
                       FROM product_attributes
                   )
               )
SELECT
    COUNT(DISTINCT fp.id) AS total_products
FROM
    final_products fp;



-- name: GetProductAttributes :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM
        categories c
    WHERE
        c.parent_id IS NULL -- Root categories

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
),
               all_active_products AS (
                   SELECT
                       p.id
                   FROM
                       products p
                   WHERE
                       p.status = 'active'
                         AND p.valid_from <= NOW()
                         AND (p.valid_to IS NULL OR p.valid_to >= NOW())
               ),
               attribute_values AS (
                   SELECT
                       pa.name AS attribute_name,
                       pav.value AS attribute_value,
                       c.position AS attribute_position
                   FROM
                       product_to_attribute_values ptav
                           JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
                           JOIN product_attributes pa ON pav.attribute_id = pa.id
                           LEFT JOIN categories c ON c.name = pav.value -- Assuming the attribute values are tied to categories and their positions
                   WHERE
                       EXISTS (
                           SELECT 1
                           FROM all_active_products
                           WHERE ptav.product_id = all_active_products.id
                       )
                   UNION ALL
                   SELECT
                       ch.name AS attribute_name,
                       c.name AS attribute_value,
                       c.position AS attribute_position
                   FROM
                       category_hierarchy ch
                           JOIN categories c ON c.parent_id = ch.id
                   WHERE
                       ch.parent_id IS NULL -- Ensuring we are only dealing with root categories and their children
               ),
               aggregated_attributes AS (
                   SELECT
                       attribute_name,
                       jsonb_agg(attribute_value ORDER BY attribute_position, attribute_value) AS aggregated_values
                   FROM
                       attribute_values
                   GROUP BY
                       attribute_name
               ),
               ordered_aggregated_attributes AS (
                   SELECT
                       aa.attribute_name,
                       aa.aggregated_values
                   FROM
                       aggregated_attributes aa
                           JOIN
                       categories c ON c.name = aa.attribute_name
                   ORDER BY
                       c.position
               )
SELECT
    jsonb_object_agg(attribute_name, aggregated_values) AS attributes,
    (SELECT COUNT(*) FROM all_active_products) AS total_products
FROM
    ordered_aggregated_attributes;

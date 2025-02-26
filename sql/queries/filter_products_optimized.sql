-- name: GetAllProductsByFiltersPriceAsc_Optimized :many
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Get total number of attribute filters (if any)
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
-- Get candidate category IDs.
-- If you have a materialized hierarchy (e.g. via ltree) consider replacing this recursion with:
--   SELECT id FROM categories WHERE path <@ (SELECT path FROM categories WHERE name = ANY(...));
category_ids AS (
  SELECT id
  FROM categories
  WHERE
      (SELECT category_keys FROM filter_input) IS NULL
   OR id IN (
        SELECT id FROM categories
        WHERE name = ANY (
          SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
        )
      )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
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
      ), 0
    )::numeric AS discount_percent,
    COALESCE(
      (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
      ), ''
    )::text AS image_url
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.*
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
          SELECT 1
          FROM product_to_attribute_values ptav
          JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
          JOIN product_attributes pa ON pav.attribute_id = pa.id
          WHERE ptav.product_id = bp.id
            AND pa.name = f.key
            AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
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
FROM filtered_products fp
ORDER BY fp.price_in_kes ASC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersPriceDesc_Optimized :many
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Get total number of attribute filters (if any)
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
-- Get candidate category IDs.
-- If you have a materialized hierarchy (e.g. via ltree) consider replacing this recursion with:
--   SELECT id FROM categories WHERE path <@ (SELECT path FROM categories WHERE name = ANY(...));
category_ids AS (
  SELECT id
  FROM categories
  WHERE
      (SELECT category_keys FROM filter_input) IS NULL
   OR id IN (
        SELECT id FROM categories
        WHERE name = ANY (
          SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
        )
      )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
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
      ), 0
    )::numeric AS discount_percent,
    COALESCE(
      (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
      ), ''
    )::text AS image_url
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.*
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
          SELECT 1
          FROM product_to_attribute_values ptav
          JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
          JOIN product_attributes pa ON pav.attribute_id = pa.id
          WHERE ptav.product_id = bp.id
            AND pa.name = f.key
            AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
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
FROM filtered_products fp
ORDER BY fp.price_in_kes DESC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersNameAsc_Optimized :many
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Get total number of attribute filters (if any)
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
-- Get candidate category IDs.
-- If you have a materialized hierarchy (e.g. via ltree) consider replacing this recursion with:
--   SELECT id FROM categories WHERE path <@ (SELECT path FROM categories WHERE name = ANY(...));
category_ids AS (
  SELECT id
  FROM categories
  WHERE
      (SELECT category_keys FROM filter_input) IS NULL
   OR id IN (
        SELECT id FROM categories
        WHERE name = ANY (
          SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
        )
      )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
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
      ), 0
    )::numeric AS discount_percent,
    COALESCE(
      (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
      ), ''
    )::text AS image_url
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.*
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
          SELECT 1
          FROM product_to_attribute_values ptav
          JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
          JOIN product_attributes pa ON pav.attribute_id = pa.id
          WHERE ptav.product_id = bp.id
            AND pa.name = f.key
            AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
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
FROM filtered_products fp
ORDER BY fp.name ASC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersNameDesc_Optimized :many
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Get total number of attribute filters (if any)
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
-- Get candidate category IDs.
-- If you have a materialized hierarchy (e.g. via ltree) consider replacing this recursion with:
--   SELECT id FROM categories WHERE path <@ (SELECT path FROM categories WHERE name = ANY(...));
category_ids AS (
  SELECT id
  FROM categories
  WHERE
      (SELECT category_keys FROM filter_input) IS NULL
   OR id IN (
        SELECT id FROM categories
        WHERE name = ANY (
          SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
        )
      )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
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
      ), 0
    )::numeric AS discount_percent,
    COALESCE(
      (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
      ), ''
    )::text AS image_url
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.*
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
          SELECT 1
          FROM product_to_attribute_values ptav
          JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
          JOIN product_attributes pa ON pav.attribute_id = pa.id
          WHERE ptav.product_id = bp.id
            AND pa.name = f.key
            AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
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
FROM filtered_products fp
ORDER BY fp.name DESC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersNewest_Optimized :many
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Get total number of attribute filters (if any)
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
-- Get candidate category IDs.
-- If you have a materialized hierarchy (e.g. via ltree) consider replacing this recursion with:
--   SELECT id FROM categories WHERE path <@ (SELECT path FROM categories WHERE name = ANY(...));
category_ids AS (
  SELECT id
  FROM categories
  WHERE (SELECT category_keys FROM filter_input) IS NULL
     OR id IN (
          SELECT id
          FROM categories
          WHERE name = ANY (
                SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
             )
       )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
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
      ), 0
    )::numeric AS discount_percent,
    COALESCE(
      (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
      ), ''
    )::text AS image_url
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.*
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
      SELECT 1
      FROM product_to_attribute_values ptav
      JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
      JOIN product_attributes pa ON pav.attribute_id = pa.id
      WHERE ptav.product_id = bp.id
        AND pa.name = f.key
        AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
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
FROM filtered_products fp
ORDER BY fp.created_at DESC
LIMIT $4 OFFSET $5;

-- name: GetAllProductsByFiltersOldest_Optimized :many
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Get total number of attribute filters (if any)
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
-- Get candidate category IDs.
-- If you have a materialized hierarchy (e.g., via ltree) consider replacing this recursion with:
--   SELECT id FROM categories WHERE path <@ (SELECT path FROM categories WHERE name = ANY(...));
category_ids AS (
  SELECT id
  FROM categories
  WHERE (SELECT category_keys FROM filter_input) IS NULL
     OR id IN (
          SELECT id
          FROM categories
          WHERE name = ANY (
                SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
             )
       )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
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
      ), 0
    )::numeric AS discount_percent,
    COALESCE(
      (
        SELECT pi.image_url
        FROM product_images pi
        WHERE pi.product_id = p.id
        ORDER BY pi.position ASC
        LIMIT 1
      ), ''
    )::text AS image_url
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.*
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
      SELECT 1
      FROM product_to_attribute_values ptav
      JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
      JOIN product_attributes pa ON pav.attribute_id = pa.id
      WHERE ptav.product_id = bp.id
        AND pa.name = f.key
        AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
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
FROM filtered_products fp
ORDER BY fp.created_at ASC
LIMIT $4 OFFSET $5;

-- name: CountAllProductsByFilters_Optimized :one
WITH filter_input AS (
  -- Split incoming JSON into category filters and attribute filters.
  -- Here we assume that category filters are those keys that match a category name.
  SELECT
    (
      SELECT jsonb_agg(key)
      FROM jsonb_each($3::jsonb)
      WHERE key IN (SELECT name FROM categories)
    ) AS category_keys,
    (
      SELECT jsonb_object_agg(key, value)
      FROM jsonb_each($3::jsonb)
      WHERE key NOT IN (SELECT name FROM categories)
    ) AS attribute_filters
),
total_attr AS (
  -- Precompute the total number of attribute filters.
  SELECT COALESCE(COUNT(*), 0) AS total_filters
  FROM jsonb_each((SELECT attribute_filters FROM filter_input))
),
category_ids AS (
  SELECT id
  FROM categories
  WHERE
      (SELECT category_keys FROM filter_input) IS NULL
   OR id IN (
        SELECT id FROM categories
        WHERE name = ANY (
          SELECT jsonb_array_elements_text((SELECT category_keys FROM filter_input))
        )
      )
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
    ), 135
  ) AS rate_to_kes
),
base_products AS (
  SELECT
    p.id
  FROM products p
  CROSS JOIN rate
  WHERE p.category_id IN (SELECT id FROM category_ids)
    AND (p.usd_price * rate.rate_to_kes)::numeric BETWEEN $1 AND $2
    AND p.status = 'active'
    AND p.valid_from <= NOW()
    AND (p.valid_to IS NULL OR p.valid_to >= NOW())
),
filtered_products AS (
  SELECT bp.id
  FROM base_products bp
  CROSS JOIN LATERAL (
    SELECT COUNT(*) AS match_count
    FROM jsonb_each((SELECT attribute_filters FROM filter_input)) AS f(key, value)
    WHERE EXISTS (
          SELECT 1
          FROM product_to_attribute_values ptav
          JOIN product_attribute_values pav ON ptav.attribute_value_id = pav.id
          JOIN product_attributes pa ON pav.attribute_id = pa.id
          WHERE ptav.product_id = bp.id
            AND pa.name = f.key
            AND pav.value = ANY (SELECT jsonb_array_elements_text(f.value))
    )
  ) a
  CROSS JOIN total_attr t
  WHERE t.total_filters = 0 OR a.match_count = t.total_filters
)
SELECT COUNT(*) AS total_products FROM filtered_products;

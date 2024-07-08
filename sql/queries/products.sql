-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock, category_id, created_by, updated_by, is_active, search_keyword)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, to_tsvector('english', $1 || ' ' || $2))
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword;

-- name: GetProductByID :one
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, stock = $5, category_id = $6, updated_at = NOW(), updated_by = $7, featured = $8, search_keyword = to_tsvector('english', $2 || ' ' || $3)
WHERE id = $1
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword;

-- name: SoftDeleteProduct :exec
UPDATE products
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: GetProductsByCategoryID :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
WHERE category_id = $1
ORDER BY name;

-- name: SearchProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
WHERE
    name ILIKE '%' || $1 || '%' OR
    description ILIKE '%' || $1 || '%' OR
    search_keyword @@ websearch_to_tsquery('english', $1)
ORDER BY ts_rank(search_keyword, websearch_to_tsquery('english', $1)) DESC;

-- name: CountProducts :one
SELECT COUNT(*) AS count
FROM products;

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
    p.*
FROM products p
WHERE p.category_id IN (SELECT ch.id FROM category_hierarchy ch)
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
WHERE p.category_id IN (SELECT ch.id FROM category_hierarchy ch);

-- name: CountFilteredProducts :one
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY($1::VARCHAR[]) -- Start with the given list of category names

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
filtered_products AS (
    SELECT
        p.id
    FROM
        products p
    LEFT JOIN
        product_sizes ps ON p.id = ps.product_id
    LEFT JOIN
        sizes s ON ps.size_id = s.id
    LEFT JOIN
        product_colors pc ON p.id = pc.product_id
    LEFT JOIN
        colors c ON pc.color_id = c.id
    WHERE
        p.category_id IN (SELECT id FROM category_hierarchy)
        -- Size filter
        AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
        -- Color filter
        AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
        AND p.price BETWEEN $2 AND $3
)
SELECT COUNT(*) AS total_count FROM filtered_products;


-- name: GetProductsByFiltersPriceAsc :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.price ASC;

-- name: GetProductsByFiltersPriceDesc :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.price DESC;

-- name: GetProductsByFiltersNameAsc :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.name ASC;

-- name: GetProductsByFiltersNameDesc :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.name DESC;

-- name: GetProductsByFiltersDefault :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.created_at DESC;

-- name: GetProductsByFiltersNewest :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.created_at DESC;

-- name: GetProductsByFiltersOldest :many
WITH RECURSIVE category_tree AS (
    SELECT c.id, c.name
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.name
    FROM categories c
    INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT
    p.id,
    p.name,
    p.description,
    p.price,
    p.stock,
    p.category_id,
    p.created_at,
    p.updated_at,
    p.is_active,
    p.created_by,
    p.updated_by,
    p.featured
FROM products p
JOIN category_tree ct ON p.category_id = ct.id
LEFT JOIN product_colors pc ON p.id = pc.product_id
LEFT JOIN colors c ON pc.color_id = c.id
LEFT JOIN product_processors pp ON p.id = pp.product_id
LEFT JOIN processors pr ON pp.processor_id = pr.id
LEFT JOIN product_storage_options pso ON p.id = pso.product_id
LEFT JOIN storage_options so ON pso.storage_option_id = so.id
LEFT JOIN product_sizes psz ON p.id = psz.product_id
LEFT JOIN sizes s ON psz.size_id = s.id
WHERE (array_length($2::text[], 1) IS NULL OR ct.name = ANY($2))
  AND (array_length($3::text[], 1) IS NULL OR c.color_name = ANY($3))
  AND (array_length($4::text[], 1) IS NULL OR pr.name = ANY($4))
  AND (array_length($5::text[], 1) IS NULL OR so.name = ANY($5))
  AND (array_length($6::text[], 1) IS NULL OR s.size = ANY($6))
  AND p.price BETWEEN $7 AND $8
ORDER BY p.created_at ASC;

-- name: GetFilterOptions :many
SELECT DISTINCT
    col.color_name AS filter_option, 'color' AS filter_type
FROM
    colors col
        JOIN
    product_colors pc ON col.id = pc.color_id

UNION

SELECT DISTINCT
    sz.size AS filter_option, 'size' AS filter_type
FROM
    sizes sz
        JOIN
    product_sizes ps ON sz.id = ps.size_id

UNION

SELECT DISTINCT
    pr.name AS filter_option, 'processor' AS filter_type
FROM
    processors pr
        JOIN
    product_processors pp ON pr.id = pp.processor_id

UNION

SELECT DISTINCT
    so.name AS filter_option, 'storage' AS filter_type
FROM
    storage_options so
        JOIN
    product_storage_options pso ON so.id = pso.storage_option_id;

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
            INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       p.price,
                       p.stock,
                       p.category_id,
                       p.created_at,
                       p.updated_at,
                       p.is_active,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       s.size,
                       c.color_name,
                       pr.name AS processor_name,
                       so.name AS storage_name
                   FROM
                       products p
                           LEFT JOIN
                       product_sizes ps ON p.id = ps.product_id
                           LEFT JOIN
                       sizes s ON ps.size_id = s.id
                           LEFT JOIN
                       product_colors pc ON p.id = pc.product_id
                           LEFT JOIN
                       colors c ON pc.color_id = c.id
                           LEFT JOIN
                       product_processors pp ON p.id = pp.product_id
                           LEFT JOIN
                       processors pr ON pp.processor_id = pr.id
                           LEFT JOIN
                       product_storage_options pso ON p.id = pso.product_id
                           LEFT JOIN
                       storage_options so ON pso.storage_option_id = so.id
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
                     AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
                     AND (pr.name = ANY($8::VARCHAR[]) OR $8 IS NULL)
                     AND (so.name = ANY($9::VARCHAR[]) OR $9 IS NULL)
                     AND p.price BETWEEN $2 AND $3
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price,
    fp.stock,
    fp.category_id,
    fp.created_at,
    fp.updated_at,
    fp.is_active,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.size,
    fp.color_name,
    fp.processor_name,
    fp.storage_name
FROM
    filtered_products fp
ORDER BY
    fp.price ASC
LIMIT
    $6 OFFSET $7;

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
            INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       p.price,
                       p.stock,
                       p.category_id,
                       p.created_at,
                       p.updated_at,
                       p.is_active,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       s.size,
                       c.color_name,
                       pr.name AS processor_name,
                       so.name AS storage_name
                   FROM
                       products p
                           LEFT JOIN
                       product_sizes ps ON p.id = ps.product_id
                           LEFT JOIN
                       sizes s ON ps.size_id = s.id
                           LEFT JOIN
                       product_colors pc ON p.id = pc.product_id
                           LEFT JOIN
                       colors c ON pc.color_id = c.id
                           LEFT JOIN
                       product_processors pp ON p.id = pp.product_id
                           LEFT JOIN
                       processors pr ON pp.processor_id = pr.id
                           LEFT JOIN
                       product_storage_options pso ON p.id = pso.product_id
                           LEFT JOIN
                       storage_options so ON pso.storage_option_id = so.id
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
                     AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
                     AND (pr.name = ANY($8::VARCHAR[]) OR $8 IS NULL)
                     AND (so.name = ANY($9::VARCHAR[]) OR $9 IS NULL)
                     AND p.price BETWEEN $2 AND $3
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price,
    fp.stock,
    fp.category_id,
    fp.created_at,
    fp.updated_at,
    fp.is_active,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.size,
    fp.color_name,
    fp.processor_name,
    fp.storage_name
FROM
    filtered_products fp
ORDER BY
    fp.price DESC
LIMIT
    $6 OFFSET $7;

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
            INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       p.price,
                       p.stock,
                       p.category_id,
                       p.created_at,
                       p.updated_at,
                       p.is_active,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       s.size,
                       c.color_name,
                       pr.name AS processor_name,
                       so.name AS storage_name
                   FROM
                       products p
                           LEFT JOIN
                       product_sizes ps ON p.id = ps.product_id
                           LEFT JOIN
                       sizes s ON ps.size_id = s.id
                           LEFT JOIN
                       product_colors pc ON p.id = pc.product_id
                           LEFT JOIN
                       colors c ON pc.color_id = c.id
                           LEFT JOIN
                       product_processors pp ON p.id = pp.product_id
                           LEFT JOIN
                       processors pr ON pp.processor_id = pr.id
                           LEFT JOIN
                       product_storage_options pso ON p.id = pso.product_id
                           LEFT JOIN
                       storage_options so ON pso.storage_option_id = so.id
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
                     AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
                     AND (pr.name = ANY($8::VARCHAR[]) OR $8 IS NULL)
                     AND (so.name = ANY($9::VARCHAR[]) OR $9 IS NULL)
                     AND p.price BETWEEN $2 AND $3
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price,
    fp.stock,
    fp.category_id,
    fp.created_at,
    fp.updated_at,
    fp.is_active,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.size,
    fp.color_name,
    fp.processor_name,
    fp.storage_name
FROM
    filtered_products fp
ORDER BY
    fp.name
LIMIT
    $6 OFFSET $7;


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
                INNER JOIN
            category_hierarchy ch ON c.parent_id = ch.id
    ),
                   filtered_products AS (
                       SELECT
                           p.id,
                           p.name,
                           p.description,
                           p.price,
                           p.stock,
                           p.category_id,
                           p.created_at,
                           p.updated_at,
                           p.is_active,
                           p.created_by,
                           p.updated_by,
                           p.featured,
                           s.size,
                           c.color_name,
                           pr.name AS processor_name,
                           so.name AS storage_name
                       FROM
                           products p
                               LEFT JOIN
                           product_sizes ps ON p.id = ps.product_id
                               LEFT JOIN
                           sizes s ON ps.size_id = s.id
                               LEFT JOIN
                           product_colors pc ON p.id = pc.product_id
                               LEFT JOIN
                           colors c ON pc.color_id = c.id
                               LEFT JOIN
                           product_processors pp ON p.id = pp.product_id
                               LEFT JOIN
                           processors pr ON pp.processor_id = pr.id
                               LEFT JOIN
                           product_storage_options pso ON p.id = pso.product_id
                               LEFT JOIN
                           storage_options so ON pso.storage_option_id = so.id
                       WHERE
                           (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                         AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
                         AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
                         AND (pr.name = ANY($8::VARCHAR[]) OR $8 IS NULL)
                         AND (so.name = ANY($9::VARCHAR[]) OR $9 IS NULL)
                         AND p.price BETWEEN $2 AND $3
                   )
    SELECT
        fp.id,
        fp.name,
        fp.description,
        fp.price,
        fp.stock,
        fp.category_id,
        fp.created_at,
        fp.updated_at,
        fp.is_active,
        fp.created_by,
        fp.updated_by,
        fp.featured,
        fp.size,
        fp.color_name,
        fp.processor_name,
        fp.storage_name
    FROM
        filtered_products fp
ORDER BY
    fp.name DESC
LIMIT
    $6 OFFSET $7;


-- name: GetAllProductsByFiltersNewest :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
    WHERE
        c.name = ANY(COALESCE($1::VARCHAR[], ARRAY[]::VARCHAR[])) OR $1 IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id
    FROM
        categories c
            INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       p.price,
                       p.stock,
                       p.category_id,
                       p.created_at,
                       p.updated_at,
                       p.is_active,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       s.size,
                       c.color_name,
                       pr.name AS processor_name,
                       so.name AS storage_name
                   FROM
                       products p
                           LEFT JOIN
                       product_sizes ps ON p.id = ps.product_id
                           LEFT JOIN
                       sizes s ON ps.size_id = s.id
                           LEFT JOIN
                       product_colors pc ON p.id = pc.product_id
                           LEFT JOIN
                       colors c ON pc.color_id = c.id
                           LEFT JOIN
                       product_processors pp ON p.id = pp.product_id
                           LEFT JOIN
                       processors pr ON pp.processor_id = pr.id
                           LEFT JOIN
                       product_storage_options pso ON p.id = pso.product_id
                           LEFT JOIN
                       storage_options so ON pso.storage_option_id = so.id
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (s.size = ANY(COALESCE($4::VARCHAR[], ARRAY[]::VARCHAR[])) OR $4 IS NULL)
                     AND (c.color_name = ANY(COALESCE($5::VARCHAR[], ARRAY[]::VARCHAR[])) OR $5 IS NULL)
                     AND (pr.name = ANY(COALESCE($8::VARCHAR[], ARRAY[]::VARCHAR[])) OR $8 IS NULL)
                     AND (so.name = ANY(COALESCE($9::VARCHAR[], ARRAY[]::VARCHAR[])) OR $9 IS NULL)
                     AND p.price BETWEEN $2 AND $3
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price,
    fp.stock,
    fp.category_id,
    fp.created_at,
    fp.updated_at,
    fp.is_active,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.size,
    fp.color_name,
    fp.processor_name,
    fp.storage_name
FROM
    filtered_products fp
ORDER BY
    fp.created_at DESC
LIMIT
    $6 OFFSET $7;


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
            INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
               filtered_products AS (
                   SELECT
                       p.id,
                       p.name,
                       p.description,
                       p.price,
                       p.stock,
                       p.category_id,
                       p.created_at,
                       p.updated_at,
                       p.is_active,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       s.size,
                       c.color_name,
                       pr.name AS processor_name,
                       so.name AS storage_name
                   FROM
                       products p
                           LEFT JOIN
                       product_sizes ps ON p.id = ps.product_id
                           LEFT JOIN
                       sizes s ON ps.size_id = s.id
                           LEFT JOIN
                       product_colors pc ON p.id = pc.product_id
                           LEFT JOIN
                       colors c ON pc.color_id = c.id
                           LEFT JOIN
                       product_processors pp ON p.id = pp.product_id
                           LEFT JOIN
                       processors pr ON pp.processor_id = pr.id
                           LEFT JOIN
                       product_storage_options pso ON p.id = pso.product_id
                           LEFT JOIN
                       storage_options so ON pso.storage_option_id = so.id
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
                     AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
                     AND (pr.name = ANY($8::VARCHAR[]) OR $8 IS NULL)
                     AND (so.name = ANY($9::VARCHAR[]) OR $9 IS NULL)
                     AND p.price BETWEEN $2 AND $3
               )
SELECT
    fp.id,
    fp.name,
    fp.description,
    fp.price,
    fp.stock,
    fp.category_id,
    fp.created_at,
    fp.updated_at,
    fp.is_active,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.size,
    fp.color_name,
    fp.processor_name,
    fp.storage_name
FROM
    filtered_products fp
ORDER BY
    fp.created_at
LIMIT
    $6 OFFSET $7;

-- name: GetTotalProductsByFilters :one
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
            INNER JOIN
        category_hierarchy ch ON c.parent_id = ch.id
),
               filtered_products AS (
                   SELECT
                       p.id
                   FROM
                       products p
                           LEFT JOIN
                       product_sizes ps ON p.id = ps.product_id
                           LEFT JOIN
                       sizes s ON ps.size_id = s.id
                           LEFT JOIN
                       product_colors pc ON p.id = pc.product_id
                           LEFT JOIN
                       colors c ON pc.color_id = c.id
                           LEFT JOIN
                       product_processors pp ON p.id = pp.product_id
                           LEFT JOIN
                       processors pr ON pp.processor_id = pr.id
                           LEFT JOIN
                       product_storage_options pso ON p.id = pso.product_id
                           LEFT JOIN
                       storage_options so ON pso.storage_option_id = so.id
                   WHERE
                       (p.category_id IN (SELECT id FROM category_hierarchy) OR $1 IS NULL)
                     AND (s.size = ANY($4::VARCHAR[]) OR $4 IS NULL)
                     AND (c.color_name = ANY($5::VARCHAR[]) OR $5 IS NULL)
                     AND (pr.name = ANY($6::VARCHAR[]) OR $6 IS NULL)
                     AND (so.name = ANY($7::VARCHAR[]) OR $7 IS NULL)
                     AND p.price BETWEEN $2 AND $3
               )
SELECT
    COUNT(*) AS total_products
FROM
    filtered_products;



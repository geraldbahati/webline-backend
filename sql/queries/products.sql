-- name: CreateProduct :one
INSERT INTO products (
    name,
    description,
    price,
    stock,
    status,
    category_id,
    created_by,
    part_number,
    updated_by
) VALUES (
             $1, $2, $3, $4, $5, $6, $7, $8, $9
         ) RETURNING
    id,
    name,
    description,
    price,
    stock,
    category_id,
    created_at,
    updated_at,
    status,
    created_by,
    updated_by,
    featured,
    search_keyword,
    slug;

-- name: GetProductByID :one
SELECT id, name, description, price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword, slug
FROM products
WHERE id = $1;

-- name: GetProductBySlug :one
SELECT p.id, p.name, p.description, p.price, p.stock, c.name AS category_name, p.created_at, p.updated_at, p.status, p.created_by, p.updated_by, p.featured, p.search_keyword, p.slug
FROM products p
         JOIN categories c ON p.category_id = c.id
WHERE p.slug = $1;


-- name: ListProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword, slug
FROM products
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, stock = $5, category_id = $6, updated_at = NOW(), updated_by = $7, featured = $8, status = $9
WHERE id = $1
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword, slug;

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
SELECT id, name, description, price, stock, category_id, created_at, updated_at, status, created_by, updated_by, featured, search_keyword
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
    p.price,
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
    p.id, p.name, p.description, p.price, p.stock, p.created_at, p.updated_at, p.status,
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
                     -- Price range filter
                     AND p.price BETWEEN $2 AND $3
                     -- Status filter
                     AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
  AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
    AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
    AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
    AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
    AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
    AND p.status = 'active'
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
    p.status,
    p.created_by,
    p.updated_by,
    p.featured,
    p.slug
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
    AND p.status = 'active'
ORDER BY p.created_at ASC;

-- name: GetFilterOptions :many
SELECT DISTINCT
    col.color_name AS filter_option, 'color' AS filter_type
FROM
    colors col
        JOIN product_colors pc ON col.id = pc.color_id
        JOIN products p ON pc.product_id = p.id
WHERE
    p.status = 'active'

UNION

SELECT DISTINCT
    sz.size AS filter_option, 'size' AS filter_type
FROM
    sizes sz
        JOIN product_sizes ps ON sz.id = ps.size_id
        JOIN products p ON ps.product_id = p.id
WHERE
    p.status = 'active'

UNION

SELECT DISTINCT
    pr.name AS filter_option, 'processor' AS filter_type
FROM
    processors pr
        JOIN product_processors pp ON pr.id = pp.processor_id
        JOIN products p ON pp.product_id = p.id
WHERE
    p.status = 'active'

UNION

SELECT DISTINCT
    so.name AS filter_option, 'storage' AS filter_type
FROM
    storage_options so
        JOIN product_storage_options pso ON so.id = pso.storage_option_id
        JOIN products p ON pso.product_id = p.id
WHERE
    p.status = 'active';

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
                       p.status,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       p.slug,
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
                    AND p.status = 'active'
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
    fp.status,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.slug,
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
                       p.status,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       p.slug,
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
                    AND p.status = 'active'
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
    fp.status,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.slug,
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
                       p.status,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       p.slug,
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
                    AND p.status = 'active'
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
    fp.status,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.slug,
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
                           p.status,
                           p.created_by,
                           p.updated_by,
                           p.featured,
                           p.slug,
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
                        AND p.status = 'active'
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
        fp.status,
        fp.created_by,
        fp.updated_by,
        fp.featured,
        fp.slug,
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
                       p.status,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       p.slug,
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
                    AND p.status = 'active'
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
    fp.status,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.slug,
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
                       p.status,
                       p.created_by,
                       p.updated_by,
                       p.featured,
                       p.slug,
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
                    AND p.status = 'active'
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
    fp.status,
    fp.created_by,
    fp.updated_by,
    fp.featured,
    fp.slug,
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
                    AND p.status = 'active'
               )
SELECT
    COUNT(*) AS total_products
FROM
    filtered_products;

-- name: GetV2Products :many
WITH first_image AS (
    SELECT DISTINCT ON (product_id) product_id, image_url
    FROM product_images
    ORDER BY product_id, created_at
)
SELECT
    p.name,
    p.price,
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
    ):: int AS totalSales,
    p.part_number AS partNumber
FROM products p
         LEFT JOIN first_image fi ON fi.product_id = p.id
         LEFT JOIN discounts d ON d.product_id = p.id AND d.start_date <= NOW() AND d.end_date >= NOW()
ORDER BY p.created_at DESC;

-- name: GetV2ProductDetailBySlug :one
WITH product_cte AS (
    SELECT
        p.id,
        p.name,
        p.description,
        p.price,
        p.slug,
        p.stock,
        p.part_number,
        p.category_id,
        p.status
    FROM
        products p
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
             json_agg(json_build_object('url', pi.image_url)) AS images
         FROM
             product_images pi
                 JOIN product_cte p ON pi.product_id = p.id
         GROUP BY
             pi.product_id
     )
SELECT
    p.id,
    p.name,
    p.description,
    p.slug,
    CAST(p.price AS FLOAT) AS price,
    CAST(p.stock AS INTEGER) AS stock,
    p.part_number,
    p.category_id,
    p.status,
    COALESCE(s.specifications, '[]'::json) AS specifications,
    COALESCE(i.images, '[]'::json) AS images
FROM
    product_cte p
        LEFT JOIN
    specs_cte s ON p.id = s.product_id
        LEFT JOIN
    images_cte i ON p.id = i.product_id;


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

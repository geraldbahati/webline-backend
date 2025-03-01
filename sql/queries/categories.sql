-- name: CreateCategory :exec
INSERT INTO categories (name, parent_id, image_url, description, meta_description, meta_title)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateCategoryImage :one
UPDATE categories
SET image_url = $2
WHERE id = $1
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position, image_url;


-- name: GetCategoryByID :one
-- Optimized to include more fields and better comments
SELECT
    id,
    name,
    parent_id,
    created_at,
    updated_at,
    is_active,
    position,
    image_url,
    slug,
    path,
    description,
    meta_title,
    meta_description
FROM categories
WHERE id = $1 AND is_active = true;

-- name: ListCategories :many
-- Optimized to include slug and path, with option to filter active only
SELECT
    id,
    name,
    parent_id,
    created_at,
    updated_at,
    is_active,
    position,
    image_url,
    slug,
    path
FROM categories
WHERE (CASE WHEN $1::boolean THEN is_active = true ELSE true END)
ORDER BY position;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, parent_id = $3, position = $4, image_url = $5,
    description = $6, meta_description = $7, meta_title = $8,
    last_updated_by = $9,
    updated_at = NOW()
WHERE id = $1
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position, image_url;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET is_active = FALSE
WHERE id = $1;

-- name: HardDeleteCategory :exec
DELETE FROM categories
WHERE id = $1;

-- name: GetCategoriesByParentID :many
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
FROM categories
WHERE parent_id = $1
ORDER BY position;

-- name: GetCategoriesWithProductsCount :many
SELECT c.id, c.name, c.parent_id, c.position, c.created_at, c.updated_at, c.is_active, c.image_url, COUNT(p.id) as products_count
FROM categories c
         LEFT JOIN products p ON c.id = p.category_id
GROUP BY c.id, c.name, c.parent_id, c.position, c.created_at, c.updated_at, c.is_active, c.image_url
ORDER BY c.position;

-- name: GetCategoryTree :many
SELECT
    id,
    name,
    parent_id,
    created_at,
    updated_at,
    is_active,
    position,
    image_url,
    nlevel(path) AS level
FROM
    categories
ORDER BY
    position;


-- name: CheckCategoryExistence :one
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE id = $1
) AS exists;

-- name: GetCategoriesWithSubcategoryCount :many
SELECT c.id, c.name, c.parent_id, c.created_at, c.position, c.updated_at, c.is_active, c.image_url, COUNT(sc.id) as subcategory_count
FROM categories c
LEFT JOIN categories sc ON c.id = sc.parent_id
GROUP BY c.id, c.name, c.parent_id, c.created_at, c.position, c.updated_at, c.is_active, c.image_url
ORDER BY c.position;

-- name: GetParentCategories :many
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
FROM categories
WHERE parent_id IS NULL
ORDER BY position;

-- name: GetCategoryByName :one
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
FROM categories
WHERE name = $1;

-- name: GetV2CategoryHierarchy :many
-- Optimized to use proper CTEs and more efficient joins
WITH
product_counts AS (
    SELECT
        category_id,
        COUNT(*) AS product_count
    FROM
        products p
    WHERE
        p.is_active = true
    GROUP BY
        category_id
),
subcategory_product_sums AS (
    SELECT
        c.id,
        SUM(pc.product_count) as descendant_product_count
    FROM
        categories c
    JOIN
        product_counts pc ON pc.category_id IN (
            SELECT c2.id FROM categories c2 WHERE c2.path <@ c.path AND c2.id != c.id
        )
    GROUP BY
        c.id
)
SELECT
    c.id,
    c.name,
    c.parent_id,
    c.is_active,
    c.position,
    c.slug,
    c.path,
    nlevel(c.path) AS level,
    COALESCE(pc.product_count, 0) AS direct_product_count,
    COALESCE(sps.descendant_product_count, 0) AS descendant_product_count,
    COALESCE(pc.product_count, 0) + COALESCE(sps.descendant_product_count, 0) AS total_products
FROM
    categories c
LEFT JOIN
    product_counts pc ON c.id = pc.category_id
LEFT JOIN
    subcategory_product_sums sps ON c.id = sps.id
WHERE
    (CASE WHEN $1::boolean THEN c.is_active = true ELSE true END)
ORDER BY
    c.path, c.position;


SELECT id, generate_ltree_path(id)
FROM categories;

-- name: GetCategoryBySlug :one
SELECT id
FROM categories
WHERE slug = $1;

-- name: GetCategoryDetailsBySlug :one
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url, description, meta_description, meta_title, slug
FROM categories
WHERE slug = $1;

-- name: GetCategorySEOBySlug :one
SELECT meta_title, meta_description
FROM categories
WHERE slug = $1;

-- name: GetDirectChildrenCategories :many
-- Gets only direct children of a category (one level down)
SELECT
    id,
    name,
    parent_id,
    created_at,
    updated_at,
    is_active,
    position,
    image_url,
    slug,
    path,
    description
FROM categories
WHERE
    parent_id = $1
    AND (CASE WHEN $2::boolean THEN is_active = true ELSE true END)
ORDER BY position;

-- name: GetAllChildrenCategories :many
-- Gets all descendants of a category using ltree path
-- Added parameter to include or exclude inactive categories
SELECT
    c.id,
    c.name,
    c.parent_id,
    c.created_at,
    c.updated_at,
    c.is_active,
    c.position,
    c.image_url,
    c.slug,
    c.path,
    c.description,
    nlevel(c.path) - nlevel(subpath(c.path, 0, nlevel((SELECT path FROM categories WHERE categories.id = $1)))) AS depth
FROM categories c
WHERE
    c.path <@ (SELECT path FROM categories WHERE categories.id = $1)
    AND c.id != $1
    AND (CASE WHEN $2::boolean THEN c.is_active = true ELSE true END)
ORDER BY c.path, c.position;

-- name: GetAllChildrenCategoriesWithProductCount :many
-- Optimized for better performance
WITH category_tree AS (
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.created_at,
        c.updated_at,
        c.is_active,
        c.position,
        c.image_url,
        c.slug,
        c.path,
        c.description,
        nlevel(c.path) - nlevel(subpath(c.path, 0, nlevel((SELECT path FROM categories WHERE categories.id = $1)))) AS depth
    FROM categories c
    WHERE
        c.path <@ (SELECT path FROM categories WHERE categories.id = $1)
        AND c.id != $1
        AND (CASE WHEN $2::boolean THEN c.is_active = true ELSE true END)
),
product_counts AS (
    SELECT
        category_id,
        COUNT(*) as product_count
    FROM products
    WHERE
        category_id IN (SELECT id FROM category_tree)
        AND is_active = true
    GROUP BY category_id
)
SELECT
    ct.id,
    ct.name,
    ct.parent_id,
    ct.created_at,
    ct.updated_at,
    ct.is_active,
    ct.position,
    ct.image_url,
    ct.slug,
    ct.path,
    ct.description,
    ct.depth,
    COALESCE(pc.product_count, 0) as product_count
FROM
    category_tree ct
LEFT JOIN
    product_counts pc ON ct.id = pc.category_id
ORDER BY
    ct.path, ct.position;

-- name: GetChildrenCategoriesByDepth :many
-- Get categories at a specific depth level from the parent
-- Useful for navigation menus with specific levels
SELECT
    c.id,
    c.name,
    c.parent_id,
    c.created_at,
    c.updated_at,
    c.is_active,
    c.position,
    c.image_url,
    c.slug,
    c.path,
    c.description,
    nlevel(c.path) - nlevel(subpath(c.path, 0, nlevel((SELECT path FROM categories WHERE categories.id = $1)))) AS depth
FROM categories c
WHERE
    c.path <@ (SELECT path FROM categories WHERE categories.id = $1)
    AND c.id != $1
    AND c.is_active = true
    AND nlevel(c.path) - nlevel(subpath(c.path, 0, nlevel((SELECT path FROM categories WHERE categories.id = $1)))) = $2
ORDER BY c.position;

-- name: BulkUpdateCategoryPositions :exec
-- Update positions of multiple categories in a single statement
UPDATE categories
SET position = t.new_position
FROM (
    SELECT
        UNNEST($1::uuid[]) as id,
        UNNEST($2::int[]) as new_position
) AS t
WHERE categories.id = t.id;

-- name: GetCategoryAncestors :many
-- Get all ancestors of a category in order from root to parent
SELECT
    a.id,
    a.name,
    a.parent_id,
    a.created_at,
    a.updated_at,
    a.is_active,
    a.position,
    a.image_url,
    a.slug,
    a.path,
    a.description,
    nlevel(a.path) AS level
FROM
    categories a
JOIN
    categories c ON a.path @> c.path AND a.id != c.id
WHERE
    c.id = $1
    AND (CASE WHEN $2::boolean THEN a.is_active = true ELSE true END)
ORDER BY
    nlevel(a.path);

-- name: SearchCategories :many
-- Search categories by name with wildcard support
SELECT
    id,
    name,
    parent_id,
    created_at,
    updated_at,
    is_active,
    position,
    image_url,
    slug,
    path,
    description
FROM
    categories
WHERE
    (name ILIKE '%' || $1 || '%' OR $1 = '')
    AND (CASE WHEN $2::boolean THEN is_active = true ELSE true END)
ORDER BY
    position
LIMIT $3 OFFSET $4;

-- name: GetCategoryStats :one
-- Get useful stats for a category including counts
SELECT
    (SELECT COUNT(*) FROM categories c1 WHERE c1.parent_id = $1) AS direct_children_count,
    (SELECT COUNT(*) FROM categories c2 WHERE c2.path <@ (SELECT path FROM categories c3 WHERE c3.id = $1) AND c2.id != $1) AS all_descendants_count,
    (SELECT COUNT(*) FROM products p1 WHERE p1.category_id = $1) AS direct_product_count,
    (
        SELECT COUNT(DISTINCT p2.id) FROM products p2
        JOIN categories c4 ON p2.category_id = c4.id
        WHERE c4.path <@ (SELECT path FROM categories c5 WHERE c5.id = $1)
    ) AS total_product_count,
    (SELECT nlevel(path) FROM categories c6 WHERE c6.id = $1) AS depth_level;

-- name: GetPopularCategories :many
-- Get categories with the most products
SELECT
    c.id,
    c.name,
    c.parent_id,
    c.created_at,
    c.updated_at,
    c.is_active,
    c.position,
    c.image_url,
    c.slug,
    c.path,
    COUNT(p.id) AS product_count
FROM
    categories c
JOIN
    products p ON c.id = p.category_id
WHERE
    c.is_active = true
GROUP BY
    c.id
ORDER BY
    product_count DESC
LIMIT $1;

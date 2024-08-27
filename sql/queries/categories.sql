-- name: CreateCategory :exec
INSERT INTO categories (name, parent_id, image_url, description, meta_description, meta_title)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: UpdateCategoryImage :one
UPDATE categories
SET image_url = $2
WHERE id = $1
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position, image_url;


-- name: GetCategoryByID :one
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
FROM categories
WHERE id = $1;

-- name: ListCategories :many
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
FROM categories
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
SELECT c.id, c.name, c.parent_id, c.created_at, c.position ,c.updated_at, c.is_active, c.image_url,COUNT(sc.id) as subcategory_count
FROM categories c
         LEFT JOIN categories sc ON c.id = sc.parent_id
GROUP BY c.id
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
WITH product_counts AS (
    SELECT
        category_id,
        COUNT(*) AS product_count
    FROM
        products
    GROUP BY
        category_id
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
    COALESCE(pc.product_count, 0) +
    COALESCE((
                 SELECT SUM(pc2.product_count)
                 FROM product_counts pc2
                 WHERE pc2.category_id IN (
                     SELECT id
                     FROM categories
                     WHERE path <@ c.path
                 )
             ), 0) AS total_products
FROM
    categories c
        LEFT JOIN
    product_counts pc ON c.id = pc.category_id
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
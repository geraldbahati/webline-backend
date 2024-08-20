-- name: CreateCategory :one
INSERT INTO categories (name, parent_id, position)
VALUES ($1, $2, $3)
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position, image_url;

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
SET name = $2, parent_id = $3, position = $4, updated_at = NOW()
WHERE id = $1
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position, image_url;

-- name: SoftDeleteCategory :exec
DELETE FROM categories
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
WITH RECURSIVE category_tree AS (
    SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
    FROM categories
    WHERE parent_id IS NULL
    UNION
    SELECT c.id, c.name, c.parent_id, c.created_at, c.updated_at, c.is_active, c.position, c.image_url
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT id, name, parent_id, created_at, updated_at, is_active, position, image_url
FROM category_tree
ORDER BY position, parent_id, name;

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
WITH RECURSIVE category_hierarchy AS (
    SELECT
        id,
        name,
        parent_id,
        position,
        ARRAY[]::uuid[] AS path,
        1 AS level
    FROM
        categories
    WHERE
        parent_id IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position,
        ch.path || c.parent_id,
        ch.level + 1
    FROM
        categories c
            INNER JOIN
        category_hierarchy ch ON ch.id = c.parent_id
)
SELECT
    id,
    name,
    parent_id,
    position,
    path,
    level
FROM
    category_hierarchy
ORDER BY
    path, level, position;

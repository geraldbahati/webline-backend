-- name: CreateCategory :one
INSERT INTO categories (name, parent_id)
VALUES ($1, $2)
    RETURNING id, name, parent_id, created_at, updated_at, is_active;


-- name: GetCategoryByID :one
SELECT id, name, parent_id, created_at, updated_at, is_active
FROM categories
WHERE id = $1;

-- name: ListCategories :many
SELECT id, name, parent_id, created_at, updated_at, is_active
FROM categories
ORDER BY name;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, parent_id = $3, updated_at = NOW()
WHERE id = $1
    RETURNING id, name, parent_id, created_at, updated_at, is_active;

-- name: SoftDeleteCategory :exec
UPDATE categories
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: GetCategoriesByParentID :many
SELECT id, name, parent_id, created_at, updated_at, is_active
FROM categories
WHERE parent_id = $1
ORDER BY name;

-- name: GetCategoriesWithProductsCount :many
SELECT c.id, c.name, c.parent_id, c.created_at, c.updated_at, c.is_active,COUNT(p.id) as products_count
FROM categories c
         LEFT JOIN products p ON c.id = p.category_id
GROUP BY c.id
ORDER BY c.name;

-- name: GetCategoryTree :many
WITH RECURSIVE category_tree AS (
    SELECT id, name, parent_id, created_at, updated_at, is_active
    FROM categories
    WHERE parent_id IS NULL
    UNION
    SELECT c.id, c.name, c.parent_id, c.created_at, c.updated_at, c.is_active
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT id, name, parent_id, created_at, updated_at, is_active
FROM category_tree
ORDER BY parent_id, name;


-- name: CheckCategoryExistence :one
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE id = $1
) AS exists;

-- name: GetCategoriesWithSubcategoryCount :many
SELECT c.id, c.name, c.parent_id, c.created_at, c.updated_at, c.is_active,COUNT(sc.id) as subcategory_count
FROM categories c
         LEFT JOIN categories sc ON c.id = sc.parent_id
GROUP BY c.id
ORDER BY c.name;

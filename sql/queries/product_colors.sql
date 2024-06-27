-- name: CreateProductColor :one
INSERT INTO product_colors (product_id, color_id)
VALUES ($1, (SELECT id FROM colors WHERE color_name = $2))
RETURNING id, product_id, color_id, created_at, updated_at;

-- name: GetProductColorByID :one
SELECT pc.id, pc.product_id, c.color_name, c.color_value, pc.created_at, pc.updated_at
FROM product_colors pc
JOIN colors c ON pc.color_id = c.id
WHERE pc.id = $1;

-- name: ListProductColorsByProductID :many
SELECT pc.id, pc.product_id, c.color_name, c.color_value, pc.created_at, pc.updated_at
FROM product_colors pc
JOIN colors c ON pc.color_id = c.id
WHERE pc.product_id = $1
ORDER BY pc.created_at;

-- name: UpdateProductColor :one
UPDATE product_colors
SET color_id = (SELECT colors.id FROM colors WHERE color_name = $2), updated_at = NOW()
WHERE product_colors.id = $1
RETURNING id, product_id, color_id, created_at, updated_at;

-- name: DeleteProductColor :exec
DELETE FROM product_colors
WHERE id = $1;

-- name: GetAvailableColorsByParentCategoryID :many
WITH RECURSIVE category_tree AS (
    SELECT c.id
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT c.color_name, c.color_value
FROM products p
JOIN product_colors pc ON p.id = pc.product_id
JOIN colors c ON pc.color_id = c.id
JOIN category_tree ct ON p.category_id = ct.id
ORDER BY c.color_name;

-- name: GetAllColors :many
SELECT DISTINCT c.id, c.color_name, c.color_value
FROM colors c
ORDER BY c.color_name;

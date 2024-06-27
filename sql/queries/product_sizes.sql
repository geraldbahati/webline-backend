-- name: GetProductSizeByID :one
SELECT ps.id, ps.product_id, s.size, ps.additional_price, ps.created_at, ps.updated_at
FROM product_sizes ps
JOIN sizes s ON ps.size_id = s.id
WHERE ps.id = $1;

-- name: ListProductSizesByProductID :many
SELECT ps.id, ps.product_id, s.size, ps.additional_price, ps.created_at, ps.updated_at
FROM product_sizes ps
JOIN sizes s ON ps.size_id = s.id
WHERE ps.product_id = $1
ORDER BY ps.created_at;

-- name: CreateProductSize :one
INSERT INTO product_sizes (product_id, size_id, additional_price)
VALUES ($1, (SELECT id FROM sizes WHERE size = $2), $3)
RETURNING id, product_id, size_id, additional_price, created_at, updated_at;

-- name: UpdateProductSize :one
UPDATE product_sizes
SET size_id = (SELECT sizes.id FROM sizes WHERE size = $2), additional_price = $3, updated_at = NOW()
WHERE product_sizes.id = $1
RETURNING id, product_id, size_id, additional_price, created_at, updated_at;

-- name: DeleteProductSize :exec
DELETE FROM product_sizes
WHERE id = $1;

-- name: GetAvailableSizesByParentCategoryID :many
WITH RECURSIVE category_tree AS (
    SELECT c.id
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT s.size
FROM products p
JOIN product_sizes ps ON p.id = ps.product_id
JOIN sizes s ON ps.size_id = s.id
JOIN category_tree ct ON p.category_id = ct.id
ORDER BY s.size;

-- name: GetAllSizes :many
SELECT DISTINCT s.id, s.size
FROM sizes s
ORDER BY s.size;

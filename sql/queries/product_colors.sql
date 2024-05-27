-- name: CreateProductColor :one
INSERT INTO product_colors (product_id, color_name)
VALUES ($1, $2)
    RETURNING id, product_id, color_name, created_at, updated_at;

-- name: GetProductColorByID :one
SELECT id, product_id, color_name, created_at, updated_at
FROM product_colors
WHERE id = $1;

-- name: ListProductColorsByProductID :many
SELECT id, product_id, color_name, created_at, updated_at
FROM product_colors
WHERE product_id = $1
ORDER BY created_at;

-- name: UpdateProductColor :one
UPDATE product_colors
SET color_name = $2, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, color_name, created_at, updated_at;

-- name: DeleteProductColor :exec
DELETE FROM product_colors
WHERE id = $1;
-- name: CreateProductImage :one
INSERT INTO product_images (product_id, image_url)
VALUES ($1, $2)
    RETURNING id, product_id, image_url, created_at, updated_at;

-- name: GetProductImageByID :one
SELECT id, product_id, image_url, created_at, updated_at
FROM product_images
WHERE id = $1;

-- name: ListProductImagesByProductID :many
SELECT id, product_id, image_url, created_at, updated_at
FROM product_images
WHERE product_id = $1
ORDER BY created_at;

-- name: UpdateProductImage :one
UPDATE product_images
SET image_url = $2, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, image_url, created_at, updated_at;

-- name: DeleteProductImage :exec
DELETE FROM product_images
WHERE id = $1;

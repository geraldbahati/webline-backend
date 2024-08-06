-- name: CreateProductImage :one
INSERT INTO product_images (product_id, image_url)
VALUES ($1, $2)
    RETURNING id, product_id, image_url, created_at, updated_at, position;

-- name: GetProductImageByID :one
SELECT id, product_id, image_url, created_at, updated_at, position
FROM product_images
WHERE id = $1;

-- name: ListProductImagesByProductID :many
SELECT id, product_id, image_url, created_at, updated_at, position
FROM product_images
WHERE product_id = $1
ORDER BY position;

-- name: UpdateProductImage :one
UPDATE product_images
SET image_url = $2, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, image_url, created_at, updated_at, position;

-- name: DeleteProductImage :exec
DELETE FROM product_images
WHERE id = $1;

-- name: GetImageKeysByProductID :many
SELECT image_url FROM product_images WHERE product_id = $1;

-- name: UpdateProductImages :exec
WITH updated_images AS (
    SELECT pi.id,
           pi.image_url,
           row_number() OVER (ORDER BY array_position($2::text[], pi.image_url)) AS new_position
    FROM product_images pi
    WHERE pi.product_id = $1 AND pi.image_url = ANY($2::text[])
)
UPDATE product_images
SET position = updated_images.new_position
FROM updated_images
WHERE product_images.id = updated_images.id;

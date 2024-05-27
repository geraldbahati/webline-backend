-- name: CreateProductOption :one
INSERT INTO product_options (product_id, option_name)
VALUES ($1, $2)
    RETURNING id, product_id, option_name, created_at, updated_at;

-- name: GetProductOptionByID :one
SELECT id, product_id, option_name, created_at, updated_at
FROM product_options
WHERE id = $1;

-- name: ListProductOptionsByProductID :many
SELECT id, product_id, option_name, created_at, updated_at
FROM product_options
WHERE product_id = $1
ORDER BY created_at;

-- name: UpdateProductOption :one
UPDATE product_options
SET option_name = $2, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, option_name, created_at, updated_at;

-- name: DeleteProductOption :exec
DELETE FROM product_options
WHERE id = $1;

-- name: CreateProductOptionValue :one
INSERT INTO product_option_values (option_id, value_name, additional_price)
VALUES ($1, $2, $3)
    RETURNING id, option_id, value_name, additional_price, created_at, updated_at;

-- name: GetProductOptionValueByID :one
SELECT id, option_id, value_name, additional_price, created_at, updated_at
FROM product_option_values
WHERE id = $1;

-- name: ListProductOptionValuesByOptionID :many
SELECT id, option_id, value_name, additional_price, created_at, updated_at
FROM product_option_values
WHERE option_id = $1
ORDER BY created_at;

-- name: UpdateProductOptionValue :one
UPDATE product_option_values
SET value_name = $2, additional_price = $3, updated_at = NOW()
WHERE id = $1
    RETURNING id, option_id, value_name, additional_price, created_at, updated_at;

-- name: DeleteProductOptionValue :exec
DELETE FROM product_option_values
WHERE id = $1;
-- name: CreateProductVariant :one
INSERT INTO product_variants (product_id, variant_name, variant_value, additional_price)
VALUES ($1, $2, $3, $4)
    RETURNING id, product_id, variant_name, variant_value, additional_price, created_at, updated_at;

-- name: GetProductVariantByID :one
SELECT id, product_id, variant_name, variant_value, additional_price, created_at, updated_at
FROM product_variants
WHERE id = $1;

-- name: ListProductVariantsByProductID :many
SELECT id, product_id, variant_name, variant_value, additional_price, created_at, updated_at
FROM product_variants
WHERE product_id = $1
ORDER BY variant_name;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET variant_name = $2, variant_value = $3, additional_price = $4, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, variant_name, variant_value, additional_price, created_at, updated_at;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants
WHERE id = $1;
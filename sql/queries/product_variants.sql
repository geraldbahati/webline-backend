-- name: CreateProductVariant :one
INSERT INTO product_variants (product_id, variant_name, variant_value, price, stock)
VALUES ($1, $2, $3, $4, $5)
    RETURNING id, product_id, variant_name, variant_value, created_at, updated_at, price, stock;

-- name: GetProductVariantByID :one
SELECT id, product_id, variant_name, variant_value, created_at, updated_at, price, stock
FROM product_variants
WHERE id = $1;

-- name: ListProductVariantsByProductID :many
SELECT id, product_id, variant_name, variant_value, created_at, updated_at , price, stock
FROM product_variants
WHERE product_id = $1
ORDER BY variant_name;

-- name: UpdateProductVariant :one
UPDATE product_variants
SET variant_name = $2, variant_value = $3, price = $4, stock = $5, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, variant_name, variant_value, created_at, updated_at, price, stock;

-- name: DeleteProductVariant :exec
DELETE FROM product_variants
WHERE id = $1;
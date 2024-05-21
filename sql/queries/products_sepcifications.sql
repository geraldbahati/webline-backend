-- name: CreateProductSpecification :one
INSERT INTO product_specifications (product_id, spec_name, spec_value)
VALUES ($1, $2, $3)
    RETURNING id, product_id, spec_name, spec_value, created_at, updated_at;

-- name: GetProductSpecificationByID :one
SELECT id, product_id, spec_name, spec_value, created_at, updated_at
FROM product_specifications
WHERE id = $1;

-- name: ListProductSpecificationsByProductID :many
SELECT id, product_id, spec_name, spec_value, created_at, updated_at
FROM product_specifications
WHERE product_id = $1
ORDER BY spec_name;

-- name: UpdateProductSpecification :one
UPDATE product_specifications
SET spec_name = $2, spec_value = $3, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, spec_name, spec_value, created_at, updated_at;

-- name: DeleteProductSpecification :exec
DELETE FROM product_specifications
WHERE id = $1;

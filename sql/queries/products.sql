-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock, category_id, created_by, updated_by, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by;

-- name: GetProductByID :one
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by
FROM products
ORDER BY name;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, stock = $5, category_id = $6, updated_at = NOW(), updated_by = $7
WHERE id = $1
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by;

-- name: SoftDeleteProduct :exec
UPDATE products
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: GetProductsByCategoryID :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by
FROM products
WHERE category_id = $1
ORDER BY name;

-- name: SearchProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by
FROM products
WHERE name ILIKE '%' || $1 || '%'
ORDER BY name;

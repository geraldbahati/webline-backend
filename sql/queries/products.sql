-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock, category_id, created_by, updated_by, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE)
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured;

-- name: GetProductByID :one
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured
FROM products
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, stock = $5, category_id = $6, updated_at = NOW(), updated_by = $7, featured = $8
WHERE id = $1
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured;

-- name: SoftDeleteProduct :exec
UPDATE products
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: GetProductsByCategoryID :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured
FROM products
WHERE category_id = $1
ORDER BY name;

-- name: SearchProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured
FROM products
WHERE (name ILIKE '%' || $1 || '%' OR description ILIKE '%' || $1 || '%')
ORDER BY name;

-- name: CountProducts :one
SELECT COUNT(*) AS count
FROM products;

-- name: GetProductsByParentCategoryID :many
WITH RECURSIVE category_tree AS (
    SELECT c.id
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.created_at, p.updated_at, p.is_active, p.created_by, p.updated_by, p.featured
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
ORDER BY p.name;

-- name: GetFilteredProductsByParentCategoryID :many
WITH RECURSIVE category_tree AS (
    SELECT c.id
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.created_at, p.updated_at, p.is_active, p.created_by, p.updated_by, p.featured
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id
        JOIN categories c2 on c2.id = p.category_id
         LEFT JOIN product_colors pc ON p.id = pc.product_id

WHERE ($2::VARCHAR[] IS NULL OR c2.name ILIKE ANY ($2::VARCHAR[]))
--   AND ($3 IS NULL OR pc.color_name ILIKE ANY ($3::VARCHAR[]))
--   WHERE ($2::DECIMAL IS NULL OR p.price >= $2::DECIMAL)
--   AND ($3::DECIMAL IS NULL OR p.price <= $3::DECIMAL)
ORDER BY p.name;




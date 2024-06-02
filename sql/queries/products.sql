-- name: CreateProduct :one
INSERT INTO products (name, description, price, stock, category_id, created_by, updated_by, is_active, search_keyword)
VALUES ($1, $2, $3, $4, $5, $6, $7, TRUE, to_tsvector('english', $1 || ' ' || $2))
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword;

-- name: GetProductByID :one
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
WHERE id = $1;

-- name: ListProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
ORDER BY name
LIMIT $1 OFFSET $2;

-- name: UpdateProduct :one
UPDATE products
SET name = $2, description = $3, price = $4, stock = $5, category_id = $6, updated_at = NOW(), updated_by = $7, featured = $8, search_keyword = to_tsvector('english', $2 || ' ' || $3)
WHERE id = $1
    RETURNING id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword;

-- name: SoftDeleteProduct :exec
UPDATE products
SET is_active = FALSE, updated_at = NOW()
WHERE id = $1;

-- name: GetProductsByCategoryID :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
WHERE category_id = $1
ORDER BY name;

-- name: SearchProducts :many
SELECT id, name, description, price, stock, category_id, created_at, updated_at, is_active, created_by, updated_by, featured, search_keyword
FROM products
WHERE
    search_keyword @@ plainto_tsquery('english', $1)
ORDER BY ts_rank(search_keyword, plainto_tsquery('english', $1)) DESC;

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
SELECT p.id, p.name, p.description, p.price, p.stock, p.category_id, p.created_at, p.updated_at, p.is_active, p.created_by, p.updated_by, p.featured, p.search_keyword
FROM products p
         JOIN category_tree ct ON p.category_id = ct.id;
ORDER BY p.name;

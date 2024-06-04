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
    name ILIKE '%' || $1 || '%' OR
    description ILIKE '%' || $1 || '%' OR
    search_keyword @@ websearch_to_tsquery('english', $1)
ORDER BY ts_rank(search_keyword, websearch_to_tsquery('english', $1)) DESC;

-- name: CountProducts :one
SELECT COUNT(*) AS count
FROM products;

-- name: GetProductsByParentCategoryID :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT
        c2.id,
        c2.name,
        c2.parent_id,
        c2.position
    FROM categories c2
             INNER JOIN category_hierarchy ch ON c2.parent_id = ch.id
)
SELECT
    p.*
FROM products p
WHERE p.category_id IN (SELECT ch.id FROM category_hierarchy ch);

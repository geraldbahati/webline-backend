 -- id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
 -- product_id UUID REFERENCES products(id) ON DELETE CASCADE,
 -- size VARCHAR(50) NOT NULL,
 -- additional_price DECIMAL(10, 2) DEFAULT 0,
 -- created_at TIMESTAMPTZ DEFAULT NOW(),
 -- updated_at TIMESTAMPTZ DEFAULT NOW()



-- name: GetProductSizeByID :one
SELECT id, product_id, size, additional_price, created_at, updated_at
FROM product_sizes
WHERE id = $1;

-- name: ListProductSizesByProductID :many
SELECT id, product_id, size, additional_price, created_at, updated_at
FROM product_sizes
WHERE product_id = $1
ORDER BY created_at;

-- name: CreateProductSize :one
INSERT INTO product_sizes (product_id, size, additional_price)
VALUES ($1, $2, $3)
    RETURNING id, product_id, size, additional_price, created_at, updated_at;

-- name: UpdateProductSize :one
UPDATE product_sizes
SET size = $2, additional_price = $3, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, size, additional_price, created_at, updated_at;

-- name: DeleteProductSize :exec
DELETE FROM product_sizes
WHERE id = $1;

-- name: GetAvailableSizesByParentCategoryID :many
WITH RECURSIVE category_tree AS (
    SELECT c.id
    FROM categories c
    WHERE c.id = $1
    UNION ALL
    SELECT c.id
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT DISTINCT ps.size
FROM products p
         JOIN product_sizes ps ON p.id = ps.product_id
         JOIN category_tree ct ON p.category_id = ct.id
ORDER BY ps.size;

-- name: GetAllSizes :many
SELECT DISTINCT id, size
FROM product_sizes
ORDER BY size;

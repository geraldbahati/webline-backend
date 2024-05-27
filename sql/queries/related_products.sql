-- name: CreateRelatedProduct :one
INSERT INTO related_products (product_id, related_product_id)
VALUES ($1, $2)
    RETURNING product_id, related_product_id;

-- name: GetRelatedProductByProductID :one
SELECT product_id, related_product_id
FROM related_products
WHERE product_id = $1;

-- name: ListRelatedProductsByProductID :many
SELECT product_id, related_product_id
FROM related_products
WHERE product_id = $1
ORDER BY related_product_id;

-- name: DeleteRelatedProduct :exec
DELETE FROM related_products
WHERE product_id = $1 AND related_product_id = $2;


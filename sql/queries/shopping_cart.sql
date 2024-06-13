-- Create a new shopping cart
-- name: CreateShoppingCart :one
INSERT INTO shopping_carts (id, user_id, session_id, total_items, total_price, created_at, updated_at)
VALUES (gen_random_uuid(), $1, gen_random_uuid(), 0, 0.0, NOW(), NOW())
RETURNING id, user_id, session_id, total_items, total_price, created_at, updated_at;

-- Get a shopping cart by user ID
-- name: GetShoppingCartByUserID :one
SELECT id, user_id, session_id, total_items, total_price, created_at, updated_at
FROM shopping_carts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- Delete a shopping cart
-- name: DeleteShoppingCart :exec
DELETE FROM shopping_carts
WHERE id = $1;

-- Get a shopping cart by session ID
-- name: GetShoppingCartBySessionID :one
SELECT id, user_id, session_id, total_items, total_price, created_at, updated_at
FROM shopping_carts
WHERE session_id = $1
ORDER BY created_at DESC
LIMIT 1;

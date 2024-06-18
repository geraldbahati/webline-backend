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

-- name: UpdateCartTotals :exec
UPDATE shopping_carts sc
SET total_price = (SELECT COALESCE(SUM(ci.price * ci.quantity), 0)
    FROM cart_items ci
    WHERE ci.shopping_cart_id = sc.id),
    total_items = (SELECT COALESCE(SUM(ci.quantity), 0)
                   FROM cart_items ci
                   WHERE ci.shopping_cart_id = sc.id)
WHERE sc.id = $1;

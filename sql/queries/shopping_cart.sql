-- Create a new shopping cart
-- name: CreateShoppingCart :one
INSERT INTO shopping_carts (user_id, session_reference, total_items, total_price)
VALUES ($1, $2, 0, 0.0)
RETURNING id, user_id, session_reference, total_items, total_price, created_at, updated_at;

-- Get a shopping cart by user ID
-- name: GetShoppingCartByUserID :one
SELECT id, user_id, session_reference, total_items, total_price, created_at, updated_at
FROM shopping_carts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- Delete a shopping cart
-- name: DeleteShoppingCart :exec
DELETE FROM shopping_carts
WHERE id = $1;

-- Get a shopping cart by session ID
-- name: GetCartBySessionReference :one
SELECT * FROM shopping_carts WHERE session_reference = $1;

-- name: UpdateCartTotals :exec
UPDATE shopping_carts
SET 
    total_price = (
        SELECT COALESCE(SUM(price * quantity), 0)
        FROM cart_items
        WHERE shopping_cart_id = shopping_carts.id
    ),
    total_items = (
        SELECT COALESCE(SUM(quantity), 0)
        FROM cart_items
        WHERE shopping_cart_id = shopping_carts.id
    ),
    updated_at = NOW()
WHERE shopping_carts.id = $1;
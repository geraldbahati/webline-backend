-- name: GetCartItem :one
SELECT id, quantity FROM cart_items
WHERE shopping_cart_id = $1 AND product_id = $2;

-- Insert or update the item in the cart
-- name: UpsertCartItem :one
INSERT INTO cart_items (id, shopping_cart_id, product_id, quantity, price, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
ON CONFLICT (shopping_cart_id, product_id)
DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity
RETURNING id, shopping_cart_id, product_id, quantity, price, created_at, updated_at;

-- Remove an item from the cart
-- name: RemoveCartItem :exec
DELETE FROM cart_items
WHERE shopping_cart_id = $1 AND product_id = $2;

-- Update the quantity of an item in the cart
-- name: UpdateCartItemQuantity :exec
UPDATE cart_items
SET quantity = $3, updated_at = NOW()
WHERE shopping_cart_id = $1 AND product_id = $2;

-- Get all items in the cart
-- name: GetAllCartItems :many
SELECT id, shopping_cart_id, product_id, quantity, price, created_at, updated_at
FROM cart_items
WHERE shopping_cart_id = $1;

-- Calculate the total price of items in the cart
-- name: CalculateCartTotal :one
SELECT SUM(quantity * price) AS total_price
FROM cart_items
WHERE shopping_cart_id = $1;

-- Remove all items from the cart
-- name: ClearCart :exec
DELETE FROM cart_items
WHERE shopping_cart_id = $1;

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

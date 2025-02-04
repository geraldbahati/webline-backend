-- cart_queries.sql

-- Get a cart item
-- name: GetCartItem :one
SELECT
    ci.id,
    ci.product_id,
    p.name,
    p.description,
    ci.quantity,
    ci.price,
    pi.image_url
FROM cart_items ci
JOIN products p ON ci.product_id = p.id
LEFT JOIN product_images pi ON p.id = pi.product_id AND (pi.position = 1 OR pi.position IS NULL)
WHERE ci.shopping_cart_id = $1 AND ci.product_id = $2
LIMIT 1;

-- Insert or update the item in the cart
-- name: UpsertCartItem :one
WITH upserted AS (
    INSERT INTO cart_items (id, shopping_cart_id, product_id, quantity, price, created_at, updated_at)
    VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
    ON CONFLICT (shopping_cart_id, product_id)
    DO UPDATE SET quantity = cart_items.quantity + EXCLUDED.quantity, updated_at = NOW()
    RETURNING id, shopping_cart_id, product_id, quantity, price, created_at, updated_at
)
SELECT
    upserted.id,
    upserted.product_id,
    products.name,
    products.description,
    upserted.quantity,
    upserted.price,
    product_images.image_url
FROM upserted
JOIN products ON upserted.product_id = products.id
LEFT JOIN product_images ON products.id = product_images.product_id
WHERE product_images.position = 1 OR product_images.position IS NULL
LIMIT 1;

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
SELECT
    ci.id,
    ci.product_id,
    p.name,
    p.description,
    ci.quantity,
    ci.price,
    pi.image_url
FROM cart_items ci
JOIN products p ON ci.product_id = p.id
LEFT JOIN product_images pi ON p.id = pi.product_id AND (pi.position = 1 OR pi.position IS NULL)
WHERE ci.shopping_cart_id = $1
GROUP BY ci.id, p.id, pi.image_url
ORDER BY ci.created_at ASC;

-- Calculate the total price of items in the cart
-- name: CalculateCartTotal :one
SELECT COALESCE(SUM(quantity * price), 0.0)::numeric(10,2) AS total_price
FROM cart_items
WHERE shopping_cart_id = $1;

-- Remove all items from the cart
-- name: ClearCart :exec
DELETE FROM cart_items
WHERE shopping_cart_id = $1;

-- Update the cart's user ID and nullify guest_id
-- name: UpdateCartUserID :exec
UPDATE shopping_carts
SET user_id = $2,
    guest_id = NULL
WHERE id = $1;

-- Update the cart to associate with a guest ID and nullify user_id
-- name: UpdateCartGuestID :exec
UPDATE shopping_carts
SET guest_id = $2,
    user_id = NULL
WHERE id = $1;

-- Get a shopping cart by user ID
-- name: GetShoppingCartByUserID :one
SELECT *
FROM shopping_carts
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- Get a shopping cart by guest ID
-- name: GetCartByGuestID :one
SELECT *
FROM shopping_carts
WHERE guest_id = $1
LIMIT 1;

-- Create a shopping cart for a user
-- name: CreateCartForUser :one
INSERT INTO shopping_carts (user_id, total_items, total_price)
VALUES ($1, 0, 0.0)
ON CONFLICT (user_id)
DO UPDATE SET updated_at = NOW()
RETURNING *;

-- Create a shopping cart for a guest
-- name: CreateCartForGuest :one
INSERT INTO shopping_carts (guest_id, total_items, total_price)
VALUES ($1, 0, 0.0)
ON CONFLICT (guest_id)
DO UPDATE SET updated_at = NOW()
RETURNING *;

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

-- Delete a shopping cart
-- name: DeleteShoppingCart :exec
DELETE FROM shopping_carts
WHERE id = $1;

-- name: CreateShoppingCart :one
INSERT INTO shopping_carts (
    id,
    user_id,
    guest_id,
    total_items,
    total_price,
    created_at,
    updated_at
)
VALUES (
    gen_random_uuid(),
    $1,  -- user_id
    $2,  -- guest_id
    0,   -- total_items
    '0', -- total_price
    NOW(),
    NOW()
)
RETURNING *;

-- Get a shopping cart by ID
-- name: GetShoppingCartByID :one
SELECT *
FROM shopping_carts
WHERE id = $1;

-- name: GetCartByOwnerID :one
SELECT *
FROM shopping_carts
WHERE user_id = $1 OR guest_id = $1
ORDER BY created_at DESC
LIMIT 1;

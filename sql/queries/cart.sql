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
SELECT SUM(quantity * price) AS total_price
FROM cart_items
WHERE shopping_cart_id = $1;

-- Remove all items from the cart
-- name: ClearCart :exec
DELETE FROM cart_items
WHERE shopping_cart_id = $1;

-- name: UpdateCartUserID :exec
UPDATE shopping_carts
SET user_id = $2
WHERE id = $1;
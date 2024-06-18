-- name: CreateOrderItem :one
INSERT INTO order_items (id, order_id, product_id, quantity, price, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, NOW(), NOW())
RETURNING id;

-- name: GetOrderItemsByOrderId :many
SELECT id, order_id, product_id, quantity, price, created_at, updated_at
FROM order_items
WHERE order_id = $1;

-- name: CreateOrderItemOption :exec
INSERT INTO order_item_options (order_item_id, option_type, option_value, additional_price)
VALUES ($1, $2, $3, $4);

-- name: CreateOrder :one
INSERT INTO orders (user_id, guest_checkout_id, status, payment_status, total)
VALUES ($1, $2, 'pending', 'pending', $3)
RETURNING id;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: UpdateOrderPaymentStatus :exec
UPDATE orders
SET payment_status = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetOrderById :one
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at
FROM orders
WHERE id = $1;

-- name: GetOrdersByUserId :many
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetOrdersByGuestCheckoutId :many
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at
FROM orders
WHERE guest_checkout_id = $1
ORDER BY created_at DESC;

-- name: CreateGuestCheckout :one
INSERT INTO guest_checkouts (id, email, first_name, last_name, phone, street_address, city, state, country, created_at, updated_at)
VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING id;

-- name: GetGuestCheckoutByEmail :one
SELECT id, email, first_name, last_name, phone, street_address, city, state, country, created_at, updated_at
FROM guest_checkouts
WHERE email = $1;

-- name: GetOrderIDsByUserID :many
SELECT id
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: CreatePayment :one
INSERT INTO order_payments (order_id, payment_id, payment_method_id, payment_status_id, amount)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, payment_id, payment_method_id, payment_status_id, amount, created_at;

-- name: UpdatePaymentStatus :exec
UPDATE order_payments
SET payment_status_id = $1, updated_at = NOW()
WHERE payment_id = $2;

-- name: GetPaymentsByOrderID :many
SELECT id, order_id, payment_id, payment_method_id, payment_status_id, amount, created_at
FROM order_payments
WHERE order_id = $1;

-- name: GetAllPayments :many
SELECT id, order_id, payment_id, payment_method_id, payment_status_id, amount, created_at
FROM order_payments;

-- name: UpdatePaymentID :exec
UPDATE order_payments
SET payment_id = $1, updated_at = NOW()
WHERE id = $2;

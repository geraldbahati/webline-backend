-- name: CreatePayment :one
INSERT INTO order_payments (order_id, checkout_request_id, payment_method_id, payment_status_id, amount)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, order_id, checkout_request_id, payment_method_id, payment_status_id, amount, created_at;

-- name: UpdatePaymentStatus :exec
UPDATE order_payments
SET payment_status_id = $1, amount = $2, result_code = $3, result_desc = $4
WHERE checkout_request_id = $5;

-- name: GetPaymentByOrderID :one
SELECT id, order_id, checkout_request_id, payment_method_id, payment_status_id, amount, result_code, result_desc, created_at
FROM order_payments
WHERE order_id = $1;

-- name: GetAllPayments :many
SELECT id, order_id, checkout_request_id, payment_method_id, payment_status_id, amount, result_code, result_desc, created_at
FROM order_payments;

-- name: UpdateCheckoutRequestIDByOrderID :exec
UPDATE order_payments
SET checkout_request_id = $1
WHERE order_id = $2;

-- name: GetStatusByID :one
SELECT status
FROM payment_statuses
WHERE id = $1;

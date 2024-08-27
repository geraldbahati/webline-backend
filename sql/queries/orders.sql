-- name: CreateOrder :one
INSERT INTO orders (user_id, guest_checkout_id, status, payment_status, total)
VALUES ($1, $2, 'pending', 'pending', $3)
RETURNING id;

-- name: UpdateOrderStatus :exec
UPDATE orders
SET status = $2, updated_at = NOW()
WHERE id = $1;

-- name: CancelOrder :one
UPDATE orders
SET status = 'cancelled', updated_at = NOW()
WHERE id = $1
RETURNING order_number;

-- name: UpdateOrderPaymentStatus :exec
UPDATE orders
SET payment_status = $2, updated_at = NOW()
WHERE id = $1;

-- name: GetOrderById :one
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at, order_number
FROM orders
WHERE id = $1;

-- name: GetOrdersByUserId :many
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at, order_number
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetOrdersByGuestCheckoutId :many
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at, order_number
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


-- name: GetUserOrGuestCheckoutNameByOrderID :one
SELECT
    u.first_name AS user_first_name,
    u.last_name AS user_last_name,
    u.phone_number AS user_phone_number,
    g.first_name AS guest_first_name,
    g.last_name AS guest_last_name,
    g.phone AS guest_phone
FROM orders o
LEFT JOIN users u ON o.user_id = u.id
LEFT JOIN guest_checkouts g ON o.guest_checkout_id = g.id
WHERE o.id = $1
  AND (o.user_id IS NOT NULL OR o.guest_checkout_id IS NOT NULL);

-- GetTotalRevenueByStatus gets the total revenue for orders with a specific payment status
-- name: GetTotalRevenueByStatus :one
SELECT COALESCE(SUM(amount), 0)::numeric AS total_revenue
FROM order_payments
WHERE payment_status_id = $1;

-- GetMonthlySales retrieves the total sales for each month from the order_payments table
-- name: GetMonthlySales :many
SELECT
    date_trunc('month', created_at) AS month,
    COALESCE(SUM(amount), 0)::numeric AS total_sales
FROM
    order_payments
WHERE
    payment_status_id = $1
GROUP BY
    month
ORDER BY
    month;

-- name: GetTotalRevenueForLastTwoMonths :one
WITH revenue_data AS (
    SELECT
        date_trunc('month', created_at) AS month,
        COALESCE(SUM(amount), 0)::numeric AS total_revenue
    FROM
        order_payments
    WHERE
        payment_status_id = $1
    GROUP BY
        month
    ORDER BY
        month DESC
    LIMIT 2
)
SELECT
    COALESCE((SELECT total_revenue FROM revenue_data ORDER BY month DESC LIMIT 1), 0)::numeric AS current_month_revenue,
    COALESCE((SELECT total_revenue FROM revenue_data ORDER BY month DESC LIMIT 1 OFFSET 1), 0)::numeric AS previous_month_revenue;

-- name: GetMonthlySalesForLastTwoMonths :one
WITH sales_data AS (
    SELECT
        date_trunc('month', created_at) AS month,
        COALESCE(SUM(amount), 0)::numeric AS total_sales
    FROM
        order_payments
    WHERE
        payment_status_id = $1
    GROUP BY
        month
    ORDER BY
        month DESC
    LIMIT 2
)
SELECT
    COALESCE((SELECT total_sales FROM sales_data ORDER BY month DESC LIMIT 1), 0)::numeric AS current_month_sales,
    COALESCE((SELECT total_sales FROM sales_data ORDER BY month DESC LIMIT 1 OFFSET 1), 0)::numeric AS previous_month_sales;

-- name: GetMonthlyRevenue :many
WITH sales_data AS (
    SELECT
        created_at AS month,
        COALESCE(SUM(amount), 0)::numeric AS total_sales
    FROM
        order_payments
    WHERE
        payment_status_id = $1
    GROUP BY
        month
    ORDER BY
        month DESC
    LIMIT 12
)
SELECT month, total_sales FROM sales_data;

-- name: GetSalesTrend :one
WITH sales_data AS (
    SELECT
        TO_CHAR(date_trunc('month', created_at), 'YYYY-MM') AS month,
        COALESCE(SUM(amount), 0) AS total_sales
    FROM
        order_payments
    WHERE
        payment_status_id = $1
    GROUP BY
        month
    ORDER BY
        month DESC
    LIMIT 2
)
SELECT
    COALESCE(MAX(total_sales), 0)::numeric AS current_month_sales,
    COALESCE(
            (SELECT total_sales FROM sales_data ORDER BY month DESC LIMIT 1 OFFSET 1),
            0
    )::numeric AS previous_month_sales
FROM sales_data;

-- name: GetRecentSales :many
SELECT
    COALESCE(u.first_name || ' ' || u.last_name, g.first_name || ' ' || g.last_name)::varchar AS name,
    COALESCE(u.email, g.email) AS email,
    op.amount,
    COALESCE(SUBSTRING(u.first_name, 1, 1) || SUBSTRING(u.last_name, 1, 1),
             SUBSTRING(g.first_name, 1, 1) || SUBSTRING(g.last_name, 1, 1))::varchar AS fallback
FROM
    orders o
        LEFT JOIN
    users u ON o.user_id = u.id
        LEFT JOIN
    guest_checkouts g ON o.guest_checkout_id = g.id
        JOIN
    order_payments op ON o.id = op.order_id
ORDER BY
    o.created_at DESC
LIMIT 10;

-- name: GetTotalSalesCurrentMonth :one
SELECT
    COUNT(*) AS total_sales
FROM
    orders
WHERE
    created_at >= date_trunc('month', current_date)
  AND created_at < date_trunc('month', current_date) + interval '1 month';

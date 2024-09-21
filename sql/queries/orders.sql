-- name: CreateOrder :one
INSERT INTO orders (
    id,
    user_id,
    guest_checkout_id,
    company_id,
    company_name,
    kra_pin,
    currency_code,
    subtotal,
    tax_amount,
    shipping_amount,
    discount_amount,
    grand_total,
    order_number,
    total
) VALUES (
    gen_random_uuid(),
    $1,  -- user_id
    $2,  -- guest_checkout_id
    $3,  -- company_id
    $4,  -- company_name
    $5,  -- kra_pin
    COALESCE($6::text, 'USD'), -- currency_code with default 'USD'
    $7,  -- subtotal
    $8,  -- tax_amount
    $9,  -- shipping_amount
    $10, -- discount_amount
    $11, -- grand_total
    $12,  -- order_number
    $13  -- total
) RETURNING id, order_number, created_at;

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
SELECT
    id,
    user_id,
    guest_checkout_id,
    company_id,
    company_name,
    kra_pin,
    status,
    payment_status,
    currency_code,
    subtotal,
    tax_amount,
    shipping_amount,
    discount_amount,
    grand_total,
    created_at,
    updated_at,
    order_number
FROM orders
WHERE id = $1;

-- name: GetOrdersByUserId :many
SELECT id, user_id,guest_checkout_id,  status, payment_status, total, created_at, updated_at, order_number
FROM orders
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GetOrdersByGuestCheckoutId :many
SELECT
    id,
    user_id,
    guest_checkout_id,
    company_id,
    company_name,
    kra_pin,
    status,
    payment_status,
    currency_code,
    subtotal,
    tax_amount,
    shipping_amount,
    discount_amount,
    grand_total,
    created_at,
    updated_at,
    order_number
FROM orders
WHERE guest_checkout_id = $1
ORDER BY created_at DESC;

-- name: CreateGuestCheckout :one
INSERT INTO guest_checkouts (
    id,
    email,
    first_name,
    last_name,
    phone,
    street_address,
    city,
    state,
    country,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    $1,  -- email
    $2,  -- first_name
    $3,  -- last_name
    $4,  -- phone
    $5,  -- street_address
    $6,  -- city
    $7,  -- state
    $8,  -- country
    now(),
    now()
) RETURNING id;

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
    g.phone AS guest_phone,
    o.company_name,
    o.kra_pin
FROM orders o
LEFT JOIN users u ON o.user_id = u.id
LEFT JOIN guest_checkouts g ON o.guest_checkout_id = g.id
WHERE o.id = $1;

-- GetTotalRevenueByStatus gets the total revenue for orders with a specific payment status
-- name: GetTotalRevenueByPaymentStatus :one
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
    FROM order_payments
    WHERE payment_status_id = $1
    GROUP BY month
    ORDER BY month DESC
    LIMIT 2
)
SELECT
    COALESCE(MAX(total_revenue), 0)::numeric AS current_month_revenue,
    COALESCE(
        (SELECT total_revenue FROM revenue_data ORDER BY month DESC OFFSET 1 LIMIT 1),
        0
    )::numeric AS previous_month_revenue
FROM revenue_data;

-- name: GetMonthlySalesForLastTwoMonths :one
WITH sales_data AS (
    SELECT
        date_trunc('month', created_at) AS month,
        COALESCE(SUM(amount), 0)::numeric AS total_sales
    FROM order_payments
    WHERE payment_status_id = $1
    GROUP BY month
    ORDER BY month DESC
    LIMIT 2
)
SELECT
    COALESCE(MAX(total_sales), 0)::numeric AS current_month_sales,
    COALESCE(
        (SELECT total_sales FROM sales_data ORDER BY month DESC OFFSET 1 LIMIT 1),
        0
    )::numeric AS previous_month_sales
FROM sales_data;

-- name: GetMonthlyRevenue :many
SELECT
    date_trunc('month', created_at) AS month,
    COALESCE(SUM(amount), 0)::numeric AS total_sales
FROM order_payments
WHERE payment_status_id = $1
GROUP BY month
ORDER BY month DESC
LIMIT 12;

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

-- name: UpdateOrderAmounts :one
WITH vat_rate AS (
  SELECT vat_percentage / 100.0 AS rate
  FROM settings
  WHERE id = TRUE
),
updated_values AS (
  SELECT
    $1::uuid AS id,
    $2::numeric AS subtotal,
    $3::numeric AS tax_amount,
    $4::numeric AS shipping_amount,
    $5::numeric AS discount_amount,
    vr.rate,
    ($2 * vr.rate) AS vat_amount,
    ($2 + $3 + $4 + ($2 * vr.rate) - $5) AS grand_total
  FROM vat_rate vr
)
UPDATE orders o
SET
  subtotal = uv.subtotal,
  tax_amount = uv.tax_amount,
  shipping_amount = uv.shipping_amount,
  discount_amount = uv.discount_amount,
  vat_amount = uv.vat_amount,
  grand_total = uv.grand_total,
  updated_at = now()
FROM updated_values uv
WHERE o.id = uv.id
RETURNING
  o.subtotal,
  o.tax_amount,
  o.shipping_amount,
  o.discount_amount,
  o.vat_amount,
  o.grand_total;
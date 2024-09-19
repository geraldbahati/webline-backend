-- name: CreateOrderItem :one
INSERT INTO order_items (
    id,
    order_id,
    product_id,
    product_name,
    product_sku,
    quantity,
    unit_price,
    total_price,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid(),
    $1,  -- order_id
    $2,  -- product_id
    $3,  -- product_name
    $4,  -- product_sku
    $5,  -- quantity
    $6,  -- unit_price
    $7,  -- total_price
    NOW(),
    NOW()
) RETURNING id;

-- name: GetOrderItemsByOrderId :many
SELECT
    id,
    order_id,
    product_id,
    product_name,
    product_sku,
    quantity,
    unit_price,
    total_price,
    created_at,
    updated_at
FROM order_items
WHERE order_id = $1;

-- name: CreateOrderItemOption :exec
INSERT INTO order_item_options (
    order_item_id,
    option_type,
    option_value,
    additional_price
) VALUES (
    $1,  -- order_item_id
    $2,  -- option_type
    $3,  -- option_value
    $4   -- additional_price
);

-- name: CreateOrderItems :exec
INSERT INTO order_items (
    order_id,
    product_id,
    quantity,
    product_name,
    product_sku,
    unit_price,
    discount_amount,
    total_price
)
SELECT
    unnest($1::uuid[]) AS order_id,
    unnest($2::uuid[]) AS product_id,
    unnest($3::integer[]) AS quantity,
    unnest($4::varchar[]) AS product_name,
    unnest($5::varchar[]) AS product_sku,
    unnest($6::numeric(12,2)[]) AS unit_price,
    unnest($7::numeric(12,2)[]) AS discount_amount,
    unnest($8::numeric(12,2)[]) AS total_price;

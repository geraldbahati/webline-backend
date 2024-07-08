-- name: CreateDiscount :one
INSERT INTO discounts (product_id, discount_percentage, start_date, end_date)
VALUES ($1, $2, $3, $4)
    RETURNING id, product_id, discount_percentage, start_date, end_date, created_at, updated_at;

-- name: GetDiscountByID :one
SELECT id, product_id, discount_percentage, start_date, end_date, created_at, updated_at
FROM discounts
WHERE id = $1;

-- name: ListDiscountsByProductID :many
SELECT id, product_id, discount_percentage, start_date, end_date, created_at, updated_at
FROM discounts
WHERE product_id = $1
ORDER BY created_at;

-- name: ListDiscounts :many
SELECT id, product_id, discount_percentage, start_date, end_date, created_at, updated_at
FROM discounts
ORDER BY created_at;

-- name: GetDiscountByProductID :one
SELECT id, product_id, discount_percentage, start_date, end_date, created_at, updated_at
FROM discounts
WHERE product_id = $1
ORDER BY created_at DESC
LIMIT 1;

-- name: UpdateDiscount :one
UPDATE discounts
SET discount_percentage = $2, start_date = $3, end_date = $4, updated_at = NOW()
WHERE id = $1
    RETURNING id, product_id, discount_percentage, start_date, end_date, created_at, updated_at;

-- name: DeleteDiscount :exec
DELETE FROM discounts
WHERE id = $1;
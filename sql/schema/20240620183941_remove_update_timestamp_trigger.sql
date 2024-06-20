-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- Add updated_at column to order_payments
ALTER TABLE public.order_payments
ADD COLUMN updated_at timestamp with time zone DEFAULT now();

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

-- Remove updated_at column from order_payments
ALTER TABLE public.order_payments
DROP COLUMN updated_at;

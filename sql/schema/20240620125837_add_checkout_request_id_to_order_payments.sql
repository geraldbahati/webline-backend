-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- Add new columns to the order_payments table
ALTER TABLE public.order_payments
ADD COLUMN checkout_request_id character varying(255) NOT NULL,
ADD COLUMN result_code integer,
ADD COLUMN result_desc text;

-- Ensure existing rows do not break with the addition of new columns
ALTER TABLE public.order_payments
ALTER COLUMN order_id SET NOT NULL;

-- Drop the payment_id column
ALTER TABLE public.order_payments
DROP COLUMN payment_id;

-- Add foreign key constraint for order_id if not already present
ALTER TABLE public.order_payments
ADD CONSTRAINT fk_order FOREIGN KEY (order_id) REFERENCES orders(id) ON DELETE CASCADE;

-- Indexes for quick lookups
CREATE INDEX idx_order_payments_checkout_request_id ON public.order_payments (checkout_request_id);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

-- Re-add the payment_id column
ALTER TABLE public.order_payments
ADD COLUMN payment_id character varying(255) NOT NULL;

-- Drop the new columns
ALTER TABLE public.order_payments
DROP COLUMN checkout_request_id,
DROP COLUMN result_code,
DROP COLUMN result_desc;

-- Drop foreign key constraint for order_id
ALTER TABLE public.order_payments
DROP CONSTRAINT fk_order;

-- Remove added indexes
DROP INDEX idx_order_payments_checkout_request_id;

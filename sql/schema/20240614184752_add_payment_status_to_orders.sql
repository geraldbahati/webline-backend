-- +goose Up
-- Adding payment_status to orders table

ALTER TABLE orders
    ADD COLUMN payment_status VARCHAR(50) NOT NULL DEFAULT 'pending';

-- Adding indexes for performance improvements
CREATE INDEX IF NOT EXISTS idx_order_user_id ON orders(user_id);
CREATE INDEX IF NOT EXISTS idx_order_status ON orders(status);

-- +goose Down
-- Dropping payment_status column from orders table

ALTER TABLE orders
DROP COLUMN IF EXISTS payment_status;

-- Removing indexes
DROP INDEX IF EXISTS idx_order_user_id;
DROP INDEX IF EXISTS idx_order_status;

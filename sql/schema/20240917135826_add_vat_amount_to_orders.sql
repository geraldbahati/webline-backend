-- +goose Up

ALTER TABLE orders
ADD COLUMN vat_amount numeric(12,2) NOT NULL DEFAULT 0;

-- Drop the existing check_grand_total constraint if it exists
ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_grand_total;

-- Add a new check_grand_total constraint including vat_amount
ALTER TABLE orders
ADD CONSTRAINT check_grand_total
CHECK (grand_total = (subtotal + tax_amount + shipping_amount + vat_amount - discount_amount));

-- +goose Down

ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_grand_total;

ALTER TABLE orders
DROP COLUMN IF EXISTS vat_amount;

-- Re-add the original check_grand_total constraint
ALTER TABLE orders
ADD CONSTRAINT check_grand_total
CHECK (grand_total = (subtotal + tax_amount + shipping_amount - discount_amount));

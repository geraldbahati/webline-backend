-- +goose Up
-- Add 'discount_amount' column and update 'check_total_price' constraint

BEGIN;

-- Step 1: Add 'discount_amount' column with default 0 and NOT NULL constraint
ALTER TABLE order_items
ADD COLUMN discount_amount numeric(12,2) NOT NULL DEFAULT 0;

-- Step 2: Drop the existing 'check_total_price' constraint
ALTER TABLE order_items
DROP CONSTRAINT IF EXISTS check_total_price;

-- Step 3: Add the new 'check_total_price' constraint to account for discounts
ALTER TABLE order_items
ADD CONSTRAINT check_total_price
CHECK (total_price = (quantity::numeric * unit_price) - discount_amount);

COMMIT;


-- +goose Down
-- Remove 'discount_amount' column and restore original 'check_total_price' constraint

BEGIN;

-- Step 1: Drop the updated 'check_total_price' constraint
ALTER TABLE order_items
DROP CONSTRAINT IF EXISTS check_total_price;

-- Step 2: Remove the 'discount_amount' column
ALTER TABLE order_items
DROP COLUMN IF EXISTS discount_amount;

-- Step 3: Restore the original 'check_total_price' constraint
ALTER TABLE order_items
ADD CONSTRAINT check_total_price
CHECK (total_price = (quantity::numeric * unit_price));

COMMIT;

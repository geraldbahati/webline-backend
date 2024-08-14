-- +goose Up
-- Add usd_price column to products table, migrate existing KES price data, and drop the price column

BEGIN;

-- Add usd_price column to store prices in USD
ALTER TABLE products
    ADD COLUMN usd_price NUMERIC(10, 2) NOT NULL DEFAULT 0;

-- Migrate existing KES price to USD assuming an initial exchange rate of 135 KES per USD
UPDATE products
SET usd_price = price / 135;

-- Drop the old price column since it’s no longer needed
ALTER TABLE products
    DROP COLUMN price;

COMMIT;

-- +goose Down
-- Revert the changes by restoring the price column and removing the usd_price column

BEGIN;

-- Add the price column back to store prices in KES
ALTER TABLE products
    ADD COLUMN price NUMERIC(10, 2) NOT NULL DEFAULT 0;

-- Restore KES price using the usd_price and the initial assumed rate of 135 KES per USD
UPDATE products
SET price = usd_price * 135;

-- Drop the usd_price column
ALTER TABLE products
    DROP COLUMN usd_price;

COMMIT;

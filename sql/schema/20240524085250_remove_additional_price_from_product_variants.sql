-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Remove additional_price column from product_variants table
ALTER TABLE product_variants
DROP COLUMN additional_price;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Add additional_price column back to product_variants table
ALTER TABLE product_variants
    ADD COLUMN additional_price DECIMAL(10, 2) DEFAULT 0;

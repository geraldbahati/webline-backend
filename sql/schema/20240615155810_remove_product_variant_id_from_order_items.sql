-- +goose Up
-- This section is executed when the migration is applied.

-- Remove product_variant_id column from order_items table
ALTER TABLE order_items
DROP COLUMN IF EXISTS product_variant_id;

-- +goose Down
-- This section is executed when the migration is rolled back.

-- Add product_variant_id column back to order_items table
ALTER TABLE order_items
ADD COLUMN product_variant_id UUID REFERENCES product_variants(id) ON DELETE SET NULL;

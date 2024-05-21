-- +goose Up
-- Add improvements to products schema

-- Add columns to the products table
ALTER TABLE products
    ADD COLUMN is_active BOOLEAN DEFAULT TRUE,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

-- Add unique constraints
ALTER TABLE products
    ADD CONSTRAINT unique_product_name_category UNIQUE (name, category_id);

ALTER TABLE product_variants
    ADD CONSTRAINT unique_product_variant UNIQUE (product_id, variant_name, variant_value);

ALTER TABLE product_specifications
    ADD CONSTRAINT unique_product_specification UNIQUE (product_id, spec_name);

-- +goose Down
-- Revert improvements to products schema

-- Remove unique constraints
ALTER TABLE products
DROP CONSTRAINT unique_product_name_category;

ALTER TABLE product_variants
DROP CONSTRAINT unique_product_variant;

ALTER TABLE product_specifications
DROP CONSTRAINT unique_product_specification;

-- Remove columns from the products table
ALTER TABLE products
DROP COLUMN is_active,
    DROP COLUMN created_by,
    DROP COLUMN updated_by;

-- +goose Up
-- SQL statements for the "up" migration.

BEGIN;

-- Drop redundant indexes
DROP INDEX IF EXISTS idx_products_name_description;

-- Add new composite index for price and status filtering
CREATE INDEX IF NOT EXISTS idx_products_price_status ON products (usd_price, status);

-- Add an index on parent_id for optimizing recursive queries in categories
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories (parent_id);

-- Add a composite index on (parent_id, name) for filtering and ordering
CREATE INDEX IF NOT EXISTS idx_categories_parent_id_name ON categories (parent_id, name);

-- Ensure an index on product_attributes.name if queries frequently filter by name
CREATE INDEX IF NOT EXISTS idx_product_attributes_name ON product_attributes (name);

-- Ensure an index on product_attribute_values.value if frequently used in queries
CREATE INDEX IF NOT EXISTS idx_product_attribute_values_value ON product_attribute_values (value);

-- Create GIN index on JSONB columns if applicable (uncomment if using JSONB attributes in products)
-- CREATE INDEX IF NOT EXISTS idx_products_attributes_jsonb ON products USING GIN (attributes);

COMMIT;

-- +goose Down
-- SQL statements for the "down" migration.

BEGIN;

-- Drop newly added indexes
DROP INDEX IF EXISTS idx_products_price_status;
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_parent_id_name;
DROP INDEX IF EXISTS idx_product_attributes_name;
DROP INDEX IF EXISTS idx_product_attribute_values_value;

-- Recreate dropped redundant indexes
CREATE INDEX IF NOT EXISTS idx_products_name_description ON products (name, description);

-- Optionally drop GIN index if created
-- DROP INDEX IF EXISTS idx_products_attributes_jsonb;

COMMIT;

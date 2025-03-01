-- +goose Up
-- +goose StatementBegin

-- Categories table indexes
CREATE INDEX IF NOT EXISTS idx_categories_parent_id ON categories(parent_id);
CREATE INDEX IF NOT EXISTS idx_categories_name ON categories(name);

-- Add materialized path pattern to optimize recursive category queries
ALTER TABLE categories ADD COLUMN IF NOT EXISTS path ltree;

-- Create a trigger function to maintain the path
CREATE OR REPLACE FUNCTION update_category_path() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        NEW.path = text2ltree(NEW.id::text);
    ELSE
        SELECT path || text2ltree(NEW.id::text) INTO NEW.path FROM categories WHERE id = NEW.parent_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop trigger if it exists
DROP TRIGGER IF EXISTS trig_update_category_path ON categories;

-- Create trigger
CREATE TRIGGER trig_update_category_path
    BEFORE INSERT OR UPDATE ON categories
    FOR EACH ROW EXECUTE FUNCTION update_category_path();

-- Update existing paths
WITH RECURSIVE category_paths AS (
    SELECT id, id::text AS path_text, parent_id
    FROM categories
    WHERE parent_id IS NULL

    UNION ALL

    SELECT c.id, cp.path_text || '.' || c.id::text, c.parent_id
    FROM categories c, category_paths cp
    WHERE c.parent_id = cp.id
)
UPDATE categories c
SET path = text2ltree(cp.path_text)
FROM category_paths cp
WHERE c.id = cp.id;

-- Create a GiST index on the path column for fast traversal
CREATE INDEX IF NOT EXISTS idx_categories_path ON categories USING GIST (path);
CREATE INDEX IF NOT EXISTS idx_categories_path_btree ON categories USING btree (path);

-- Products table indexes
CREATE INDEX IF NOT EXISTS idx_products_category_id ON products(category_id);
CREATE INDEX IF NOT EXISTS idx_products_status ON products(status);
CREATE INDEX IF NOT EXISTS idx_products_validity ON products(valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_products_usd_price ON products(usd_price);
CREATE INDEX IF NOT EXISTS idx_products_created_at ON products(created_at);
CREATE INDEX IF NOT EXISTS idx_products_slug ON products(slug);

-- Exchange rates table indexes
CREATE INDEX IF NOT EXISTS idx_exchange_rates_currency_validity ON exchange_rates(currency_code, valid_from, valid_to);

-- Product attributes related indexes
CREATE INDEX IF NOT EXISTS idx_product_to_attribute_values_product_id ON product_to_attribute_values(product_id);
CREATE INDEX IF NOT EXISTS idx_product_to_attribute_values_attr_val_id ON product_to_attribute_values(attribute_value_id);

CREATE INDEX IF NOT EXISTS idx_product_attribute_values_attribute_id ON product_attribute_values(attribute_id);
CREATE INDEX IF NOT EXISTS idx_product_attribute_values_value ON product_attribute_values(value);

CREATE INDEX IF NOT EXISTS idx_product_attributes_name ON product_attributes(name);

-- Discounts table indexes
CREATE INDEX IF NOT EXISTS idx_discounts_product_id ON discounts(product_id);
CREATE INDEX IF NOT EXISTS idx_discounts_date_range ON discounts(start_date, end_date);

-- Product images table indexes
CREATE INDEX IF NOT EXISTS idx_product_images_product_id_position ON product_images(product_id, position);

-- Composite indexes for common query patterns
CREATE INDEX IF NOT EXISTS idx_products_active_category ON products(category_id, status) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_products_price_filter ON products(category_id, status, usd_price) WHERE status = 'active';

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Remove trigger and function
DROP TRIGGER IF EXISTS trig_update_category_path ON categories;
DROP FUNCTION IF EXISTS update_category_path();

-- Remove path column
ALTER TABLE categories DROP COLUMN IF EXISTS path;

-- Remove all indexes created in the up migration
DROP INDEX IF EXISTS idx_categories_parent_id;
DROP INDEX IF EXISTS idx_categories_name;
DROP INDEX IF EXISTS idx_categories_path;
DROP INDEX IF EXISTS idx_categories_path_btree;

DROP INDEX IF EXISTS idx_products_category_id;
DROP INDEX IF EXISTS idx_products_status;
DROP INDEX IF EXISTS idx_products_validity;
DROP INDEX IF EXISTS idx_products_usd_price;
DROP INDEX IF EXISTS idx_products_created_at;
DROP INDEX IF EXISTS idx_products_slug;

DROP INDEX IF EXISTS idx_exchange_rates_currency_validity;

DROP INDEX IF EXISTS idx_product_to_attribute_values_product_id;
DROP INDEX IF EXISTS idx_product_to_attribute_values_attr_val_id;

DROP INDEX IF EXISTS idx_product_attribute_values_attribute_id;
DROP INDEX IF EXISTS idx_product_attribute_values_value;

DROP INDEX IF EXISTS idx_product_attributes_name;

DROP INDEX IF EXISTS idx_discounts_product_id;
DROP INDEX IF EXISTS idx_discounts_date_range;

DROP INDEX IF EXISTS idx_product_images_product_id_position;

DROP INDEX IF EXISTS idx_products_active_category;
DROP INDEX IF EXISTS idx_products_price_filter;

-- +goose StatementEnd

-- +goose Up
-- Add new columns to products table
ALTER TABLE public.products
    ADD COLUMN price_per_unit numeric(10, 2) NOT NULL DEFAULT 0,
    ADD COLUMN valid_from timestamp with time zone NOT NULL DEFAULT now(),
    ADD COLUMN valid_to timestamp with time zone;

-- Set the `price_per_unit` to the same value as `usd_price`
UPDATE public.products
SET price_per_unit = usd_price;

-- Add indexes to the new columns
CREATE INDEX idx_products_price_per_unit ON public.products (price_per_unit);
CREATE INDEX idx_products_valid_from ON public.products (valid_from);
CREATE INDEX idx_products_valid_to ON public.products (valid_to);

-- Create a trigger function to set `valid_to` based on category
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_valid_to_based_on_category()
    RETURNS trigger AS $$
DECLARE
    category_name text;
BEGIN
    -- Fetch the category name
    SELECT name INTO category_name FROM public.categories WHERE id = NEW.category_id;

    -- Determine the validity period based on the category
    IF category_name = 'Laptops' OR category_name = 'Desktops' THEN
        NEW.valid_to := NEW.valid_from + INTERVAL '3 months';
    ELSE
        NEW.valid_to := NEW.valid_from + INTERVAL '1 year';
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create a trigger to apply the function before insert or update
CREATE TRIGGER trigger_set_valid_to
    BEFORE INSERT OR UPDATE ON public.products
    FOR EACH ROW EXECUTE FUNCTION set_valid_to_based_on_category();

-- +goose Down
-- Remove the columns and associated indexes
DROP TRIGGER IF EXISTS trigger_set_valid_to ON public.products;
DROP FUNCTION IF EXISTS set_valid_to_based_on_category();
DROP INDEX IF EXISTS idx_products_valid_to;
DROP INDEX IF EXISTS idx_products_valid_from;
DROP INDEX IF EXISTS idx_products_price_per_unit;

ALTER TABLE public.products
    DROP COLUMN IF EXISTS price_per_unit,
    DROP COLUMN IF EXISTS valid_from,
    DROP COLUMN IF EXISTS valid_to;

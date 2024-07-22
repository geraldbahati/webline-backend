-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Add the slug column to products table
ALTER TABLE products ADD COLUMN slug character varying(255);

-- Ensure slug is unique
ALTER TABLE products ADD CONSTRAINT unique_product_slug UNIQUE (slug);

-- Create the function to generate slugs
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION generate_slug() RETURNS TRIGGER AS $$
DECLARE
    generated_slug text;
BEGIN
    -- Convert the name to a slug
    generated_slug := lower(regexp_replace(NEW.name, '[^a-zA-Z0-9]+', '-', 'g'));

    -- Ensure the slug is unique by appending a number if necessary
    WHILE EXISTS (SELECT 1 FROM products WHERE slug = generated_slug) LOOP
            generated_slug := generated_slug || '-' || (SELECT count(*) + 1 FROM products WHERE slug LIKE generated_slug || '%');
        END LOOP;

    NEW.slug := generated_slug;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create the trigger to set the slug before insert or update
-- +goose StatementBegin
CREATE TRIGGER set_slug_before_insert_or_update
    BEFORE INSERT OR UPDATE ON products
    FOR EACH ROW
    WHEN (NEW.slug IS NULL OR NEW.slug = '')
EXECUTE FUNCTION generate_slug();
-- +goose StatementEnd

-- Generate initial slugs for existing products
UPDATE products
SET slug = lower(regexp_replace(name, '[^a-zA-Z0-9]+', '-', 'g'))
WHERE slug IS NULL OR slug = '';

-- Ensure all slugs are unique
-- +goose StatementBegin
WITH duplicates AS (
    SELECT id, slug, ROW_NUMBER() OVER (PARTITION BY slug ORDER BY id) as rn
    FROM products
)
UPDATE products
SET slug = duplicates.slug || '-' || (duplicates.rn - 1)
FROM duplicates
WHERE products.id = duplicates.id AND duplicates.rn > 1;
-- +goose StatementEnd

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Drop the trigger and function
DROP TRIGGER IF EXISTS set_slug_before_insert_or_update ON products;
DROP FUNCTION IF EXISTS generate_slug();

-- Remove the slug column and its constraint
ALTER TABLE products DROP CONSTRAINT unique_product_slug;
ALTER TABLE products DROP COLUMN slug;

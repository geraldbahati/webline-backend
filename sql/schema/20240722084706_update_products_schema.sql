-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Make category_id non-nullable if needed
ALTER TABLE products
    ALTER COLUMN category_id SET NOT NULL;

-- Add CHECK constraint to price for positive values
ALTER TABLE products
    ADD CONSTRAINT check_positive_price CHECK (price >= 0);

-- Ensure meta_title, meta_description, and meta_keywords have appropriate defaults (if necessary)
ALTER TABLE products
    ALTER COLUMN meta_title SET DEFAULT '',
    ALTER COLUMN meta_description SET DEFAULT '',
    ALTER COLUMN meta_keywords SET DEFAULT '';

-- Add constraints to part_number if necessary
ALTER TABLE products
    ALTER COLUMN part_number SET NOT NULL,
    ADD CONSTRAINT part_number_unique UNIQUE (part_number);

-- Add a default value for search_keyword
ALTER TABLE products
    ALTER COLUMN search_keyword SET DEFAULT '';

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Revert changes (if necessary)
ALTER TABLE products
    ALTER COLUMN category_id DROP NOT NULL;
ALTER TABLE products
    DROP CONSTRAINT check_positive_price;
ALTER TABLE products
    ALTER COLUMN meta_title DROP DEFAULT,
    ALTER COLUMN meta_description DROP DEFAULT,
    ALTER COLUMN meta_keywords DROP DEFAULT;
ALTER TABLE products
    ALTER COLUMN part_number DROP NOT NULL,
    DROP CONSTRAINT part_number_unique;
ALTER TABLE products
    ALTER COLUMN search_keyword DROP DEFAULT;

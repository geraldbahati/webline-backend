-- +goose Up
-- Migration script to add colors table and modify product_colors to use color_id

BEGIN;

-- Step 1: Create the colors table
CREATE TABLE colors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    color_name VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Create the trigger function to update `updated_at`
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = now();
   RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

-- Attach the trigger to the colors table
-- +goose StatementBegin
CREATE TRIGGER update_colors_updated_at BEFORE UPDATE
ON colors FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();
-- +goose StatementEnd

-- Step 2: Add color_id column to product_colors table
ALTER TABLE product_colors ADD COLUMN color_id UUID;

-- Create the trigger function to update `updated_at` for product_colors
CREATE TRIGGER update_product_colors_updated_at BEFORE UPDATE
ON product_colors FOR EACH ROW EXECUTE PROCEDURE update_updated_at_column();

-- Step 3: Populate colors table with unique color names from product_colors
INSERT INTO colors (color_name)
SELECT DISTINCT color_name
FROM product_colors
WHERE color_name IS NOT NULL;

-- Step 4: Update product_colors to reference colors table
UPDATE product_colors pc
SET color_id = c.id
FROM colors c
WHERE pc.color_name = c.color_name;

-- Step 5: Add foreign key constraint and indexes
ALTER TABLE product_colors
    ADD CONSTRAINT fk_color_id FOREIGN KEY (color_id) REFERENCES colors (id) ON DELETE RESTRICT;

CREATE INDEX idx_product_colors_color_id ON product_colors(color_id);

-- Step 6: Remove the color_name column from product_colors
ALTER TABLE product_colors DROP COLUMN color_name;

COMMIT;

-- +goose Down
-- Reverse the migration

BEGIN;

-- Step 1: Add color_name column back to product_colors
ALTER TABLE product_colors ADD COLUMN color_name VARCHAR(50);

-- Step 2: Populate the color_name column in product_colors
UPDATE product_colors pc
SET color_name = c.color_name
FROM colors c
WHERE pc.color_id = c.id;

-- Step 3: Remove foreign key constraint and indexes
ALTER TABLE product_colors DROP CONSTRAINT fk_color_id;
DROP INDEX IF EXISTS idx_product_colors_color_id;

-- Step 4: Remove color_id column from product_colors
ALTER TABLE product_colors DROP COLUMN color_id;

-- Step 5: Drop colors table and associated trigger
DROP TRIGGER IF EXISTS update_colors_updated_at ON colors;
DROP FUNCTION IF EXISTS update_updated_at_column;
DROP TABLE IF EXISTS colors;

-- Drop the trigger for product_colors
DROP TRIGGER IF EXISTS update_product_colors_updated_at ON product_colors;

COMMIT;

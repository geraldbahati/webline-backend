-- +goose Up
-- Migration script to add color_value to colors table and modify product_colors to use color_id

BEGIN;

-- Step 1: Add color_value column to colors table
ALTER TABLE colors ADD COLUMN color_value VARCHAR(7);

-- Step 2: Populate color_value column in colors table
-- Note: You will need to manually update this query with the appropriate color values (hex) for each color name.
UPDATE colors
SET color_value = CASE color_name
    WHEN 'Black' THEN '#000000'
    WHEN 'White' THEN '#FFFFFF'
    -- Add other colors here
    ELSE '#000000' -- Default value for any unmatched color names
END;

COMMIT;

-- +goose Down
-- Reverse the migration

BEGIN;

-- Step 5: Drop color_value column from colors table
ALTER TABLE colors DROP COLUMN color_value;

COMMIT;

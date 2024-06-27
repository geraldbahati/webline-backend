-- +goose Up
-- Migration script to add sizes table and modify product_sizes to use size_id

BEGIN;

-- Step 1: Create the sizes table
CREATE TABLE sizes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    size VARCHAR(50) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

-- Step 2: Add size_id column to product_sizes table
ALTER TABLE product_sizes ADD COLUMN size_id UUID;

-- Step 3: Populate sizes table with unique sizes from product_sizes
INSERT INTO sizes (size)
SELECT DISTINCT size
FROM product_sizes
WHERE size IS NOT NULL;

-- Step 4: Update product_sizes to reference sizes table
UPDATE product_sizes ps
SET size_id = s.id
FROM sizes s
WHERE ps.size = s.size;

-- Step 5: Add foreign key constraint and indexes
ALTER TABLE product_sizes
    ADD CONSTRAINT fk_size_id FOREIGN KEY (size_id) REFERENCES sizes (id) ON DELETE RESTRICT;

CREATE INDEX idx_product_sizes_size_id ON product_sizes(size_id);

-- Step 6: Remove the size column from product_sizes
ALTER TABLE product_sizes DROP COLUMN size;

COMMIT;

-- +goose Down
-- Reverse the migration

BEGIN;

-- Step 1: Add size column back to product_sizes
ALTER TABLE product_sizes ADD COLUMN size VARCHAR(50);

-- Step 2: Populate the size column in product_sizes
UPDATE product_sizes ps
SET size = s.size
FROM sizes s
WHERE ps.size_id = s.id;

-- Step 3: Remove foreign key constraint and indexes
ALTER TABLE product_sizes DROP CONSTRAINT fk_size_id;
DROP INDEX IF EXISTS idx_product_sizes_size_id;

-- Step 4: Remove size_id column from product_sizes
ALTER TABLE product_sizes DROP COLUMN size_id;

-- Step 5: Drop sizes table
DROP TABLE IF EXISTS sizes;

COMMIT;

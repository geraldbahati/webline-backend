-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- Step 1: Add the new 'status' field with default value 'active' and ensure it's not nullable
ALTER TABLE products ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active';

-- Step 2: Add a CHECK constraint to ensure 'status' can only have specific values
ALTER TABLE products ADD CONSTRAINT status_check CHECK (status IN ('draft', 'archived', 'active'));

-- Step 3: Update existing records to set the 'status' field to 'active'
UPDATE products SET status = 'active';

-- Step 4: Remove the 'is_active' field
ALTER TABLE products DROP COLUMN is_active;

-- Step 5: Create new indexes to optimize searches
CREATE INDEX idx_products_price ON products (price);
CREATE INDEX idx_products_status ON products (status);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

-- Step 1: Remove new indexes
DROP INDEX IF EXISTS idx_products_price;
DROP INDEX IF EXISTS idx_products_status;

-- Step 2: Add the 'is_active' field back
ALTER TABLE products ADD COLUMN is_active BOOLEAN DEFAULT true;

-- Step 3: Update the 'is_active' field based on 'status' field
UPDATE products SET is_active = (status = 'active');

-- Step 4: Remove the CHECK constraint from the 'status' field
ALTER TABLE products DROP CONSTRAINT status_check;

-- Step 5: Remove the 'status' field
ALTER TABLE products DROP COLUMN status;

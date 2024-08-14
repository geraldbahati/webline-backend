-- +goose Up
-- Add the status column with a default value
ALTER TABLE promotions
    ADD COLUMN status VARCHAR(20) NOT NULL DEFAULT 'active',
    ADD CONSTRAINT status_check CHECK (status IN ('active', 'archived', 'draft'));

-- Create an index on the status column for faster queries
CREATE INDEX idx_promotions_status ON promotions(status);

-- +goose Down
-- Remove the index
DROP INDEX IF EXISTS idx_promotions_status;

-- Remove the status column and associated constraint
ALTER TABLE promotions
    DROP CONSTRAINT IF EXISTS status_check,
    DROP COLUMN IF EXISTS status;

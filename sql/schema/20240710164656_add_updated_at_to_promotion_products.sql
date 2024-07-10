-- +goose Up
-- SQL in transaction

-- Add updated_at column
ALTER TABLE promotion_products ADD COLUMN updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();


-- +goose Down
-- SQL in transaction

-- Remove updated_at column
ALTER TABLE promotion_products DROP COLUMN updated_at;

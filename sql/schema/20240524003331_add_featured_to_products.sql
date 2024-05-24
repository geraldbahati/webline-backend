-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Add featured column to products table
ALTER TABLE products ADD COLUMN featured BOOLEAN DEFAULT FALSE;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Remove featured column from products table
ALTER TABLE products DROP COLUMN featured;

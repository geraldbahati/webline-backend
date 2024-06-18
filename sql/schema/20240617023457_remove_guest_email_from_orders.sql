-- +goose Up
ALTER TABLE orders DROP COLUMN IF EXISTS guest_email;

-- +goose Down
ALTER TABLE orders ADD COLUMN guest_email VARCHAR(255);

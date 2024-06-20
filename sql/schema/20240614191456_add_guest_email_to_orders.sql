-- +goose Up
-- Adding guest_email to orders table and ensuring either user_id or guest_email is provided

-- +goose StatementBegin
ALTER TABLE orders
    ADD COLUMN guest_email VARCHAR(255);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE orders
    ADD CONSTRAINT check_user_or_guest CHECK (user_id IS NOT NULL OR guest_email IS NOT NULL);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS idx_order_guest_email ON orders(guest_email);
-- +goose StatementEnd

-- +goose Down
-- Removing guest_email from orders table and the related check constraint and index

-- +goose StatementBegin
ALTER TABLE orders
DROP COLUMN IF EXISTS guest_email;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE orders
DROP CONSTRAINT IF EXISTS check_user_or_guest;
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_order_guest_email;
-- +goose StatementEnd

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_user_or_guest;

ALTER TABLE orders ADD CONSTRAINT check_user_or_guest CHECK (user_id IS NOT NULL OR guest_checkout_id IS NOT NULL);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE orders DROP CONSTRAINT IF EXISTS check_user_or_guest;

ALTER TABLE orders ADD CONSTRAINT check_user_or_guest CHECK (user_id IS NOT NULL OR guest_email IS NOT NULL);

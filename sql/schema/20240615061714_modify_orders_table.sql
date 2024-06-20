-- +goose Up
-- Modifying orders table to support guest checkouts

-- +goose StatementBegin
ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS guest_checkout_id UUID REFERENCES guest_checkouts(id) ON DELETE CASCADE;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_user_or_guest') THEN
        ALTER TABLE orders
            ADD CONSTRAINT check_user_or_guest CHECK (user_id IS NOT NULL OR guest_checkout_id IS NOT NULL);
    END IF;
END $$;
-- +goose StatementEnd

-- +goose Down
-- Reverting changes to orders table

-- +goose StatementBegin
ALTER TABLE orders
    DROP COLUMN IF EXISTS guest_checkout_id;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_user_or_guest') THEN
        ALTER TABLE orders
            DROP CONSTRAINT check_user_or_guest;
    END IF;
END $$;
-- +goose StatementEnd

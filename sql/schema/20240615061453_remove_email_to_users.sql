-- +goose Up
-- Modifying users table: removing username and making email unique

-- +goose StatementBegin
ALTER TABLE users
    DROP COLUMN IF EXISTS username,
    ADD CONSTRAINT unique_email UNIQUE (email);
-- +goose StatementEnd


-- +goose Down
-- Reverting changes to users table: adding username and removing unique constraint on email

-- +goose StatementBegin
ALTER TABLE users
    ADD COLUMN username VARCHAR(255),
    DROP CONSTRAINT IF EXISTS unique_email;
-- +goose StatementEnd

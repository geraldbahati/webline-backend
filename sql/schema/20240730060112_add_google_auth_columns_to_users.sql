-- +goose Up
ALTER TABLE users
    ADD COLUMN provider VARCHAR(50),
    ADD COLUMN provider_id VARCHAR(255),
    ALTER COLUMN hashed_password DROP NOT NULL;

-- +goose Down
ALTER TABLE users
    DROP COLUMN provider,
    DROP COLUMN provider_id,
    ALTER COLUMN hashed_password SET NOT NULL;

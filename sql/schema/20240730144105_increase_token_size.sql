-- +goose Up
ALTER TABLE verification_tokens
    ALTER COLUMN token TYPE VARCHAR(512);

-- +goose Down
ALTER TABLE verification_tokens
    ALTER COLUMN token TYPE VARCHAR(255);

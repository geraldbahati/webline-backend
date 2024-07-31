-- +goose Up
CREATE TABLE password_reset_tokens (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                       email VARCHAR(255) NOT NULL REFERENCES users(email) ON DELETE CASCADE,
                                       token VARCHAR(255) NOT NULL UNIQUE,
                                       expires_at TIMESTAMPTZ NOT NULL,
                                       created_at TIMESTAMPTZ DEFAULT now(),
                                       CONSTRAINT password_reset_tokens_email_token_key UNIQUE (email, token)
);

-- Create indexes
CREATE INDEX idx_password_reset_tokens_email ON password_reset_tokens(email);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);

-- +goose Down
DROP INDEX IF EXISTS idx_password_reset_tokens_email;
DROP INDEX IF EXISTS idx_password_reset_tokens_expires_at;
DROP TABLE IF EXISTS password_reset_tokens;

-- +goose Up
CREATE TABLE verification_tokens (
                                     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                     email VARCHAR(255) NOT NULL,
                                     token VARCHAR(255) NOT NULL UNIQUE,
                                     expires_at TIMESTAMPTZ NOT NULL,
                                     created_at TIMESTAMPTZ DEFAULT now(),
                                     CONSTRAINT fk_user_email FOREIGN KEY (email) REFERENCES users(email) ON DELETE CASCADE,
                                     CONSTRAINT unique_email_token UNIQUE (email, token)
);

-- Indexes
CREATE INDEX idx_verification_tokens_token ON verification_tokens(token);
CREATE INDEX idx_verification_tokens_email ON verification_tokens(email);
CREATE INDEX idx_verification_tokens_expires_at ON verification_tokens(expires_at);

-- +goose Down
DROP INDEX idx_verification_tokens_token;
DROP INDEX idx_verification_tokens_email;
DROP INDEX idx_verification_tokens_expires_at;
DROP TABLE verification_tokens;

-- +goose Up
-- Create the admin_approval_tokens table
CREATE TABLE admin_approval_tokens (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                       request_id UUID NOT NULL REFERENCES admin_requests(id) ON DELETE CASCADE,
                                       token VARCHAR(512) NOT NULL UNIQUE,
                                       expires_at TIMESTAMPTZ NOT NULL,
                                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Additional indexes for admin_approval_tokens
CREATE INDEX idx_admin_approval_tokens_request_id ON admin_approval_tokens(request_id);
CREATE INDEX idx_admin_approval_tokens_expires_at ON admin_approval_tokens(expires_at);

-- +goose Down
-- Drop the admin_approval_tokens table and its indexes
DROP INDEX IF EXISTS idx_admin_approval_tokens_expires_at;
DROP INDEX IF EXISTS idx_admin_approval_tokens_request_id;
DROP TABLE IF EXISTS admin_approval_tokens;

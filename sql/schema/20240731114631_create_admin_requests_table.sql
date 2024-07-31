-- +goose Up
CREATE TABLE admin_requests (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                reason TEXT NOT NULL CHECK (char_length(reason) > 0),
                                status VARCHAR(50) NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
                                created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                                updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Partial unique index to ensure only one PENDING request per user
CREATE UNIQUE INDEX unique_pending_request ON admin_requests(user_id)
    WHERE status = 'PENDING';

-- Additional indexes
CREATE INDEX idx_admin_requests_user_id ON admin_requests(user_id);
CREATE INDEX idx_admin_requests_status ON admin_requests(status);
CREATE INDEX idx_admin_requests_created_at ON admin_requests(created_at);

-- +goose Down
DROP INDEX IF EXISTS idx_admin_requests_created_at;
DROP INDEX IF EXISTS idx_admin_requests_status;
DROP INDEX IF EXISTS idx_admin_requests_user_id;
DROP INDEX IF EXISTS unique_pending_request;
DROP TABLE IF EXISTS admin_requests;

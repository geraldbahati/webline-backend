-- +goose Up
-- +goose StatementBegin
ALTER TABLE sessions
ADD COLUMN csrf_token VARCHAR(255) NOT NULL DEFAULT '';

-- Optionally, populate existing sessions with a generated CSRF token
UPDATE sessions
SET csrf_token = gen_random_uuid()::text;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE sessions
DROP COLUMN csrf_token;
-- +goose StatementEnd

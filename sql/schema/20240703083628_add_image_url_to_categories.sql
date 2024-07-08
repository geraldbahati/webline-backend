-- +goose Up
-- +goose StatementBegin
ALTER TABLE categories ADD COLUMN image_url VARCHAR(255);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE categories DROP COLUMN image_url;
-- +goose StatementEnd

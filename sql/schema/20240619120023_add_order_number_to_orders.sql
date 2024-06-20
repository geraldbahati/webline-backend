-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
ALTER TABLE orders ADD COLUMN order_number VARCHAR(20) UNIQUE;
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
ALTER TABLE orders DROP COLUMN order_number;
-- +goose StatementEnd

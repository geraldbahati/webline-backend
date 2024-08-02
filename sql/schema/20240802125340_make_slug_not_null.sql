-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE products
    ALTER COLUMN slug SET NOT NULL;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

ALTER TABLE products
    ALTER COLUMN slug DROP NOT NULL;

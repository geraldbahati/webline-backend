-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE categories
    ADD COLUMN position INTEGER NOT NULL DEFAULT -1;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

ALTER TABLE categories
DROP COLUMN position;

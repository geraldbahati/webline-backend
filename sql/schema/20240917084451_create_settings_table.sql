-- +goose Up

CREATE TABLE settings (
    id              BOOLEAN PRIMARY KEY CHECK (id),
    vat_percentage  NUMERIC(5,2) NOT NULL DEFAULT 16.00 CHECK (vat_percentage >= 0 AND vat_percentage <= 100)
);

-- Insert the singleton row
INSERT INTO settings (id, vat_percentage) VALUES (TRUE, 16.00);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS settings;
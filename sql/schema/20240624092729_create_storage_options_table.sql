-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE storage_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_storage_options_name ON storage_options(name);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP INDEX IF EXISTS idx_storage_options_name;
DROP TABLE IF EXISTS storage_options;

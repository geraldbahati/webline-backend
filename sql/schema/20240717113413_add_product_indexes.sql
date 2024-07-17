-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

CREATE INDEX IF NOT EXISTS idx_products_search_keyword ON products USING gin(search_keyword);
CREATE INDEX IF NOT EXISTS idx_products_name_description ON products (name, description);
CREATE INDEX IF NOT EXISTS idx_products_is_active ON products (is_active);

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back.

DROP INDEX IF EXISTS idx_products_search_keyword;
DROP INDEX IF EXISTS idx_products_name_description;
DROP INDEX IF EXISTS idx_products_is_active;

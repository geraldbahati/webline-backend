-- +goose Up
-- +goose StatementBegin
ALTER TABLE products ADD COLUMN search_keyword tsvector;
-- +goose StatementEnd

-- +goose StatementBegin
-- Initialize search_keyword field with current data
UPDATE products SET search_keyword = to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''));
-- +goose StatementEnd

-- +goose StatementBegin
-- Create index on search_keyword
CREATE INDEX idx_products_search_keyword ON products USING gin(search_keyword);
-- +goose StatementEnd

-- +goose StatementBegin
-- Create a trigger function to update search_keyword
CREATE FUNCTION update_search_keyword() RETURNS trigger AS $$
BEGIN
  NEW.search_keyword := to_tsvector('english', coalesce(NEW.name, '') || ' ' || coalesce(NEW.description, ''));
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
-- Create trigger to update search_keyword on insert or update
CREATE TRIGGER trigger_update_search_keyword
    BEFORE INSERT OR UPDATE ON products
                         FOR EACH ROW EXECUTE FUNCTION update_search_keyword();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS trigger_update_search_keyword ON products;
DROP FUNCTION IF EXISTS update_search_keyword;
DROP INDEX IF EXISTS idx_products_search_keyword;
ALTER TABLE products DROP COLUMN search_keyword;
-- +goose StatementEnd

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE product_storage_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    storage_option_id UUID REFERENCES storage_options(id) ON DELETE CASCADE,
    additional_price NUMERIC(10, 2) DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

CREATE INDEX idx_product_storage_options_product_id ON product_storage_options(product_id);
CREATE INDEX idx_product_storage_options_storage_option_id ON product_storage_options(storage_option_id);

CREATE TRIGGER update_product_storage_options_timestamp
BEFORE UPDATE ON product_storage_options
FOR EACH ROW
EXECUTE PROCEDURE update_timestamp();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TRIGGER IF EXISTS update_product_storage_options_timestamp ON product_storage_options;
DROP INDEX IF EXISTS idx_product_storage_options_product_id;
DROP INDEX IF EXISTS idx_product_storage_options_storage_option_id;
DROP TABLE IF EXISTS product_storage_options;

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE product_processors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    processor_id UUID REFERENCES processors(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
);

-- +goose StatementBegin
CREATE INDEX idx_product_processors_product_id ON product_processors(product_id);
CREATE INDEX idx_product_processors_processor_id ON product_processors(processor_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER update_product_processors_timestamp
BEFORE UPDATE ON product_processors
FOR EACH ROW
EXECUTE PROCEDURE update_timestamp();
-- +goose StatementEnd


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TRIGGER IF EXISTS update_product_processors_timestamp ON product_processors;
DROP INDEX IF EXISTS idx_product_processors_product_id;
DROP INDEX IF EXISTS idx_product_processors_processor_id;
DROP TABLE IF EXISTS product_processors;

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

BEGIN;

-- Create ENUM type for attribute_type
CREATE TYPE attribute_type_enum AS ENUM ('size', 'color', 'RAM', 'storage');

-- Create product_attributes table
CREATE TABLE product_attributes (
                                    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                    name VARCHAR(100) NOT NULL,
                                    attribute_type attribute_type_enum NOT NULL,
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                    UNIQUE (name)
);

CREATE INDEX idx_product_attributes_attribute_type ON product_attributes (attribute_type);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trigger_update_product_attributes_updated_at
    BEFORE UPDATE ON product_attributes
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create product_attribute_values table
CREATE TABLE product_attribute_values (
                                          id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                          attribute_id UUID REFERENCES product_attributes(id) ON DELETE CASCADE,
                                          value VARCHAR(255) NOT NULL,
                                          category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
                                          hex_value VARCHAR(7), -- Add hex value for colors (e.g., '#FFFFFF')
                                          created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                          updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                          UNIQUE (attribute_id, value, category_id)
);

CREATE INDEX idx_product_attribute_values_attribute_id_category_id ON product_attribute_values (attribute_id, category_id);

CREATE TRIGGER trigger_update_product_attribute_values_updated_at
    BEFORE UPDATE ON product_attribute_values
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- Create product_to_attribute_values mapping table
CREATE TABLE product_to_attribute_values (
                                             id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                             product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                             attribute_value_id UUID REFERENCES product_attribute_values(id) ON DELETE CASCADE,
                                             created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                             updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                                             UNIQUE (product_id, attribute_value_id)
);

CREATE INDEX idx_product_to_attribute_values_product_id_attribute_value_id ON product_to_attribute_values (product_id, attribute_value_id);

CREATE TRIGGER trigger_update_product_to_attribute_values_updated_at
    BEFORE UPDATE ON product_to_attribute_values
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

COMMIT;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

BEGIN;

DROP TRIGGER IF EXISTS trigger_update_product_to_attribute_values_updated_at ON product_to_attribute_values;
DROP TRIGGER IF EXISTS trigger_update_product_attribute_values_updated_at ON product_attribute_values;
DROP TRIGGER IF EXISTS trigger_update_product_attributes_updated_at ON product_attributes;

DROP INDEX IF EXISTS idx_product_to_attribute_values_product_id_attribute_value_id;
DROP INDEX IF EXISTS idx_product_attribute_values_attribute_id_category_id;
DROP INDEX IF EXISTS idx_product_attributes_attribute_type;

DROP TABLE IF EXISTS product_to_attribute_values;
DROP TABLE IF EXISTS product_attribute_values;
DROP TABLE IF EXISTS product_attributes;

-- Drop ENUM type
DROP TYPE IF EXISTS attribute_type_enum;

COMMIT;

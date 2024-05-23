-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Create Product Options Table
CREATE TABLE product_options (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                 option_name VARCHAR(255) NOT NULL,
                                 created_at TIMESTAMPTZ DEFAULT NOW(),
                                 updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_product_options_product_id ON product_options(product_id);
CREATE INDEX idx_product_options_option_name ON product_options(option_name);

-- Create Product Option Values Table
CREATE TABLE product_option_values (
                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                       option_id UUID REFERENCES product_options(id) ON DELETE CASCADE,
                                       value_name VARCHAR(255) NOT NULL,
                                       additional_price DECIMAL(10, 2) DEFAULT 0,
                                       created_at TIMESTAMPTZ DEFAULT NOW(),
                                       updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_product_option_values_option_id ON product_option_values(option_id);
CREATE INDEX idx_product_option_values_value_name ON product_option_values(value_name);

-- Modify Product Variants Table
ALTER TABLE product_variants
    ADD COLUMN price DECIMAL(10, 2) NOT NULL,
ADD COLUMN stock INT DEFAULT 0;
CREATE INDEX idx_product_variants_product_id ON product_variants(product_id);
CREATE INDEX idx_product_variants_variant_name ON product_variants(variant_name);

-- Create Product Colors Table
CREATE TABLE product_colors (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                color_name VARCHAR(50) NOT NULL,
                                created_at TIMESTAMPTZ DEFAULT NOW(),
                                updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_product_colors_product_id ON product_colors(product_id);
CREATE INDEX idx_product_colors_color_name ON product_colors(color_name);

-- Create Discounts Table
CREATE TABLE discounts (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                           discount_percentage DECIMAL(5, 2) NOT NULL,
                           start_date TIMESTAMPTZ,
                           end_date TIMESTAMPTZ,
                           created_at TIMESTAMPTZ DEFAULT NOW(),
                           updated_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX idx_discounts_product_id ON discounts(product_id);
CREATE INDEX idx_discounts_start_date ON discounts(start_date);
CREATE INDEX idx_discounts_end_date ON discounts(end_date);

-- Create Related Products Table
CREATE TABLE related_products (
                                  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                  related_product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                  PRIMARY KEY (product_id, related_product_id)
);
CREATE INDEX idx_related_products_product_id ON related_products(product_id);
CREATE INDEX idx_related_products_related_product_id ON related_products(related_product_id);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Drop the new columns from Product Variants Table
ALTER TABLE product_variants
DROP COLUMN price,
DROP COLUMN stock;

-- Drop Product Option Values Table
DROP TABLE IF EXISTS product_option_values;

-- Drop Product Options Table
DROP TABLE IF EXISTS product_options;

-- Drop Product Colors Table
DROP TABLE IF EXISTS product_colors;

-- Drop Product Reviews Table
DROP TABLE IF EXISTS product_reviews;

-- Drop Discounts Table
DROP TABLE IF EXISTS discounts;

-- Drop Related Products Table
DROP TABLE IF EXISTS related_products;

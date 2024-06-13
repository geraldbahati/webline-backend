-- +goose Up
-- Create product_sizes table
CREATE TABLE product_sizes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    size VARCHAR(50) NOT NULL,
    additional_price DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for product_sizes table
CREATE INDEX idx_product_sizes_product_id ON product_sizes(product_id);
CREATE INDEX idx_product_sizes_size ON product_sizes(size);


-- +goose Down
-- Drop product_sizes table
DROP TABLE IF EXISTS product_sizes;

-- Remove indexes from product_sizes table
DROP INDEX IF EXISTS idx_product_sizes_product_id;
DROP INDEX IF EXISTS idx_product_sizes_size;

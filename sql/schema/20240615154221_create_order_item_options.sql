-- +goose Up
-- Create order_item_options table
CREATE TABLE order_item_options (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_item_id UUID REFERENCES order_items(id) ON DELETE CASCADE,
    option_type VARCHAR(50) NOT NULL, -- e.g., 'size', 'color'
    option_value VARCHAR(50) NOT NULL,
    additional_price DECIMAL(10, 2) DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create indexes for order_item_options table
CREATE INDEX idx_order_item_options_order_item_id ON order_item_options(order_item_id);
CREATE INDEX idx_order_item_options_option_type ON order_item_options(option_type);
CREATE INDEX idx_order_item_options_option_value ON order_item_options(option_value);

-- +goose Down
-- Drop order_item_options table
DROP TABLE IF EXISTS order_item_options;

-- Remove indexes from order_item_options table
DROP INDEX IF EXISTS idx_order_item_options_order_item_id;
DROP INDEX IF EXISTS idx_order_item_options_option_type;
DROP INDEX IF EXISTS idx_order_item_options_option_value;

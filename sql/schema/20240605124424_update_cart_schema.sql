-- +goose Up

-- Adding constraints and checks to existing tables
-- +goose StatementBegin
ALTER TABLE orders
    ADD CONSTRAINT check_status CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_status_history
    ADD CONSTRAINT check_status CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
    ADD CONSTRAINT check_status CHECK (status IN ('pending', 'paid', 'failed')),
    ADD CONSTRAINT check_method CHECK (method IN ('cash', 'credit_card', 'debit_card', 'paypal'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_shipments
    ADD CONSTRAINT check_status CHECK (status IN ('pending', 'shipped', 'delivered'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE shipment
    ADD CONSTRAINT check_shipment_status CHECK (shipment_status IN ('pending', 'shipped', 'delivered'));
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_items
    ADD CONSTRAINT check_quantity CHECK (quantity > 0);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE cart_items
    ADD COLUMN price DECIMAL(10, 2) NOT NULL,
    ADD CONSTRAINT check_quantity CHECK (quantity > 0);
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE shopping_carts
    ADD CONSTRAINT check_total_items CHECK (total_items >= 0),
    ADD CONSTRAINT check_total_price CHECK (total_price >= 0);
-- +goose StatementEnd

-- Creating indexes for performance improvements
-- +goose StatementBegin
CREATE INDEX idx_order_user_id ON orders(user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_order_status ON orders(status);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_order_item_order_id ON order_items(order_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_cart_user_id ON shopping_carts(user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_cart_item_cart_id ON cart_items(shopping_cart_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_wishlist_user_id ON wishlists(user_id);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_wishlist_item_wishlist_id ON wishlist_items(wishlist_id);
-- +goose StatementEnd

-- Adding triggers to auto-update timestamps
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER update_order_timestamp
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER update_order_item_timestamp
BEFORE UPDATE ON order_items
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER update_cart_item_timestamp
BEFORE UPDATE ON cart_items
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER update_wishlist_item_timestamp
BEFORE UPDATE ON wishlist_items
FOR EACH ROW
EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose Down

-- Removing triggers
-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_order_timestamp ON orders;
DROP TRIGGER IF EXISTS update_order_item_timestamp ON order_items;
DROP TRIGGER IF EXISTS update_cart_item_timestamp ON cart_items;
DROP TRIGGER IF EXISTS update_wishlist_item_timestamp ON wishlist_items;
-- +goose StatementEnd

-- Removing indexes
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_order_user_id;
DROP INDEX IF EXISTS idx_order_status;
DROP INDEX IF EXISTS idx_order_item_order_id;
DROP INDEX IF EXISTS idx_cart_user_id;
DROP INDEX IF EXISTS idx_cart_item_cart_id;
DROP INDEX IF EXISTS idx_wishlist_user_id;
DROP INDEX IF EXISTS idx_wishlist_item_wishlist_id;
-- +goose StatementEnd

-- Removing constraints and checks
-- +goose StatementBegin
ALTER TABLE orders
    DROP CONSTRAINT IF EXISTS check_status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_status_history
    DROP CONSTRAINT IF EXISTS check_status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_payments
    DROP CONSTRAINT IF EXISTS check_status,
    DROP CONSTRAINT IF EXISTS check_method;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_shipments
    DROP CONSTRAINT IF EXISTS check_status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE shipment
    DROP CONSTRAINT IF EXISTS check_shipment_status;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE order_items
    DROP CONSTRAINT IF EXISTS check_quantity;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE cart_items
    DROP CONSTRAINT IF EXISTS check_quantity,
    DROP COLUMN IF EXISTS price;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE shopping_carts
    DROP CONSTRAINT IF EXISTS check_total_items,
    DROP CONSTRAINT IF EXISTS check_total_price;
-- +goose StatementEnd

-- Dropping function
-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_timestamp();
-- +goose StatementEnd

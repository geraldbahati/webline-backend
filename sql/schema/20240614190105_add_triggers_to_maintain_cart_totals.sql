-- +goose Up
-- Adding triggers to maintain total price and total items in shopping cart

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_cart_totals()
RETURNS TRIGGER AS $$
BEGIN
UPDATE shopping_carts
SET total_price = (SELECT COALESCE(SUM(price * quantity), 0) FROM cart_items WHERE shopping_cart_id = NEW.shopping_cart_id),
    total_items = (SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE shopping_cart_id = NEW.shopping_cart_id)
WHERE id = NEW.shopping_cart_id;
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cart_items_after_insert_update_delete') THEN
CREATE TRIGGER cart_items_after_insert_update_delete
    AFTER INSERT OR UPDATE OR DELETE ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION update_cart_totals();
END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
-- Ensuring price consistency for cart items
CREATE OR REPLACE FUNCTION set_cart_item_price()
RETURNS TRIGGER AS $$
BEGIN
    NEW.price := (SELECT price FROM products WHERE id = NEW.product_id);
RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cart_items_before_insert') THEN
CREATE TRIGGER cart_items_before_insert
    BEFORE INSERT ON cart_items
    FOR EACH ROW
    EXECUTE FUNCTION set_cart_item_price();
END IF;
END $$;
-- +goose StatementEnd


-- +goose Down
-- Removing triggers and functions

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cart_items_after_insert_update_delete') THEN
DROP TRIGGER cart_items_after_insert_update_delete ON cart_items;
END IF;
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cart_items_before_insert') THEN
DROP TRIGGER cart_items_before_insert ON cart_items;
END IF;
END $$;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_cart_totals;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS set_cart_item_price;
-- +goose StatementEnd

-- +goose Up

-- Adding constraints and checks to existing tables
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_total_items') THEN
        ALTER TABLE shopping_carts ADD CONSTRAINT check_total_items CHECK (total_items >= 0);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_total_price') THEN
        ALTER TABLE shopping_carts ADD CONSTRAINT check_total_price CHECK (total_price >= 0);
    END IF;
END $$;
-- +goose StatementEnd

-- Creating indexes for performance improvements
-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_cart_user_id') THEN
        CREATE INDEX idx_cart_user_id ON shopping_carts(user_id);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_cart_item_cart_id') THEN
        CREATE INDEX idx_cart_item_cart_id ON cart_items(shopping_cart_id);
    END IF;
END $$;
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
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_cart_item_timestamp') THEN
        CREATE TRIGGER update_cart_item_timestamp
        BEFORE UPDATE ON cart_items
        FOR EACH ROW
        EXECUTE FUNCTION update_timestamp();
    END IF;
END $$;
-- +goose StatementEnd

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

-- Ensuring price consistency for cart items
-- +goose StatementBegin
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

-- Dropping triggers
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'update_cart_item_timestamp') THEN
        DROP TRIGGER update_cart_item_timestamp ON cart_items;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cart_items_after_insert_update_delete') THEN
        DROP TRIGGER cart_items_after_insert_update_delete ON cart_items;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'cart_items_before_insert') THEN
        DROP TRIGGER cart_items_before_insert ON cart_items;
    END IF;
END $$;
-- +goose StatementEnd

-- Dropping functions
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_timestamp') THEN
        DROP FUNCTION update_timestamp;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'update_cart_totals') THEN
        DROP FUNCTION update_cart_totals;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'set_cart_item_price') THEN
        DROP FUNCTION set_cart_item_price;
    END IF;
END $$;
-- +goose StatementEnd

-- Removing indexes
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_cart_user_id') THEN
        DROP INDEX idx_cart_user_id;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_class WHERE relname = 'idx_cart_item_cart_id') THEN
        DROP INDEX idx_cart_item_cart_id;
    END IF;
END $$;
-- +goose StatementEnd

-- Removing constraints and checks
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_total_items') THEN
        ALTER TABLE shopping_carts DROP CONSTRAINT check_total_items;
    END IF;
    IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'check_total_price') THEN
        ALTER TABLE shopping_carts DROP CONSTRAINT check_total_price;
    END IF;
END $$;
-- +goose StatementEnd

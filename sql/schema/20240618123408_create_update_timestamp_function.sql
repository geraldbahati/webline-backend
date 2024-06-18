-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_order_timestamp'
    ) THEN
        CREATE TRIGGER update_order_timestamp BEFORE UPDATE ON orders FOR EACH ROW EXECUTE FUNCTION update_timestamp();
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_order_item_timestamp'
    ) THEN
        CREATE TRIGGER update_order_item_timestamp BEFORE UPDATE ON order_items FOR EACH ROW EXECUTE FUNCTION update_timestamp();
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose StatementBegin
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_trigger WHERE tgname = 'update_order_payments_timestamp'
    ) THEN
        CREATE TRIGGER update_order_payments_timestamp BEFORE UPDATE ON order_payments FOR EACH ROW EXECUTE FUNCTION update_timestamp();
    END IF;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_order_timestamp ON orders;
DROP TRIGGER IF EXISTS update_order_item_timestamp ON order_items;
DROP TRIGGER IF EXISTS update_order_payments_timestamp ON order_payments;
DROP FUNCTION IF EXISTS update_timestamp;
-- +goose StatementEnd

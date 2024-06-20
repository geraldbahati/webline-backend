-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_cart_totals() RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'DELETE' THEN
        UPDATE shopping_carts
        SET total_items = (SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE shopping_cart_id = OLD.shopping_cart_id),
            total_price = (SELECT COALESCE(SUM(quantity * price), 0) FROM cart_items WHERE shopping_cart_id = OLD.shopping_cart_id),
            updated_at = NOW()
        WHERE id = OLD.shopping_cart_id;
    ELSE
        UPDATE shopping_carts
        SET total_items = (SELECT COALESCE(SUM(quantity), 0) FROM cart_items WHERE shopping_cart_id = NEW.shopping_cart_id),
            total_price = (SELECT COALESCE(SUM(quantity * price), 0) FROM cart_items WHERE shopping_cart_id = NEW.shopping_cart_id),
            updated_at = NOW()
        WHERE id = NEW.shopping_cart_id;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
DROP FUNCTION IF EXISTS update_cart_totals() CASCADE;
-- +goose StatementEnd

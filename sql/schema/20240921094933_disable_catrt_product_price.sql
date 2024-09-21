-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_cart_item_price() RETURNS TRIGGER AS $$
BEGIN
    -- Set the price of the cart item to the usd_price from the products table
    NEW.price := (SELECT usd_price FROM products WHERE id = NEW.product_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_cart_item_price() RETURNS TRIGGER AS $$
BEGIN
    NEW.price := (SELECT price FROM products WHERE id = NEW.product_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
CREATE SEQUENCE order_number_seq;

CREATE OR REPLACE FUNCTION generate_order_number() RETURNS TRIGGER AS $$
BEGIN
    NEW.order_number := TO_CHAR(NEXTVAL('order_number_seq'), 'FM000000');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER set_order_number
BEFORE INSERT ON orders
FOR EACH ROW
EXECUTE FUNCTION generate_order_number();
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
DROP TRIGGER IF EXISTS set_order_number ON orders;
DROP FUNCTION IF EXISTS generate_order_number();
DROP SEQUENCE IF EXISTS order_number_seq;
-- +goose StatementEnd

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
DROP TRIGGER IF EXISTS cart_items_after_insert_update_delete ON cart_items;
CREATE TRIGGER cart_items_after_insert_update_delete
AFTER INSERT OR DELETE OR UPDATE ON cart_items
FOR EACH ROW EXECUTE FUNCTION update_cart_totals();
-- +goose StatementEnd

-- +goose StatementBegin
DROP TRIGGER IF EXISTS cart_items_before_insert ON cart_items;
CREATE TRIGGER cart_items_before_insert
BEFORE INSERT ON cart_items
FOR EACH ROW EXECUTE FUNCTION set_cart_item_price();
-- +goose StatementEnd

-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_cart_item_timestamp ON cart_items;
CREATE TRIGGER update_cart_item_timestamp
BEFORE UPDATE ON cart_items
FOR EACH ROW EXECUTE FUNCTION update_timestamp();
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
DROP TRIGGER IF EXISTS cart_items_after_insert_update_delete ON cart_items;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TRIGGER IF EXISTS cart_items_before_insert ON cart_items;
-- +goose StatementEnd

-- +goose StatementBegin
DROP TRIGGER IF EXISTS update_cart_item_timestamp ON cart_items;
-- +goose StatementEnd

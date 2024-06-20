-- +goose Up
ALTER TABLE cart_items
    ADD CONSTRAINT unique_shopping_cart_product UNIQUE (shopping_cart_id, product_id);

-- +goose Down
ALTER TABLE cart_items
DROP CONSTRAINT unique_shopping_cart_product;

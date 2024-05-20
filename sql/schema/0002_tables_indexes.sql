-- +goose Up
-- Users Table
CREATE UNIQUE INDEX users_username_idx ON users (username);
CREATE UNIQUE INDEX users_email_idx ON users (email);
CREATE INDEX users_created_at_idx ON users (created_at);

-- Products Table
CREATE INDEX products_name_idx ON products (name);
CREATE INDEX products_category_id_idx ON products (category_id);
CREATE INDEX products_category_name_idx ON products (category_id, name); -- Composite Index

-- Categories Table
CREATE INDEX categories_name_idx ON categories (name);

-- Product Variants Table
CREATE INDEX product_variants_product_id_idx ON product_variants (product_id);

-- Product Images Table
CREATE INDEX product_images_product_id_idx ON product_images (product_id);

-- Product Specifications Table
CREATE INDEX product_specifications_product_id_idx ON product_specifications (product_id);

-- Product Reviews Table
CREATE INDEX product_reviews_product_id_idx ON product_reviews (product_id);
CREATE INDEX product_reviews_user_id_idx ON product_reviews (user_id);

-- Orders Table
CREATE INDEX orders_user_id_idx ON orders (user_id);
CREATE INDEX orders_status_idx ON orders (status);
CREATE INDEX orders_user_status_idx ON orders (user_id, status); -- Composite Index

-- Order Items Table
CREATE INDEX order_items_order_id_idx ON order_items (order_id);

-- Order Status History Table
CREATE INDEX order_status_history_order_id_idx ON order_status_history (order_id);

-- Order Payments Table
CREATE INDEX order_payments_order_id_idx ON order_payments (order_id);

-- Order Shipments Table
CREATE INDEX order_shipments_order_id_idx ON order_shipments (order_id);

-- Shopping Carts Table
CREATE INDEX shopping_carts_user_id_idx ON shopping_carts (user_id);

-- Cart Items Table
CREATE INDEX cart_items_shopping_cart_id_idx ON cart_items (shopping_cart_id);

-- Wishlists Table
CREATE INDEX wishlists_user_id_idx ON wishlists (user_id);

-- Wishlist Items Table
CREATE INDEX wishlist_items_wishlist_id_idx ON wishlist_items (wishlist_id);

-- User Addresses Table
CREATE INDEX user_addresses_user_id_idx ON user_addresses (user_id);

-- User Roles Table
CREATE INDEX user_roles_user_id_idx ON user_roles (user_id);

-- Refresh Tokens Table
CREATE INDEX refresh_tokens_user_id_idx ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_expires_at_idx ON refresh_tokens (expires_at);

-- Shipment Table
CREATE INDEX shipment_order_id_idx ON shipment (order_id);

-- +goose Down
-- Users Table
DROP INDEX IF EXISTS users_username_idx;
DROP INDEX IF EXISTS users_email_idx;
DROP INDEX IF EXISTS users_created_at_idx;

-- Products Table
DROP INDEX IF EXISTS products_name_idx;
DROP INDEX IF EXISTS products_category_id_idx;
DROP INDEX IF EXISTS products_category_name_idx;

-- Categories Table
DROP INDEX IF EXISTS categories_name_idx;

-- Product Variants Table
DROP INDEX IF EXISTS product_variants_product_id_idx;

-- Product Images Table
DROP INDEX IF EXISTS product_images_product_id_idx;

-- Product Specifications Table
DROP INDEX IF EXISTS product_specifications_product_id_idx;

-- Product Reviews Table
DROP INDEX IF EXISTS product_reviews_product_id_idx;
DROP INDEX IF EXISTS product_reviews_user_id_idx;

-- Orders Table
DROP INDEX IF EXISTS orders_user_id_idx;
DROP INDEX IF EXISTS orders_status_idx;
DROP INDEX IF EXISTS orders_user_status_idx;

-- Order Items Table
DROP INDEX IF EXISTS order_items_order_id_idx;

-- Order Status History Table
DROP INDEX IF EXISTS order_status_history_order_id_idx;

-- Order Payments Table
DROP INDEX IF EXISTS order_payments_order_id_idx;

-- Order Shipments Table
DROP INDEX IF EXISTS order_shipments_order_id_idx;

-- Shopping Carts Table
DROP INDEX IF EXISTS shopping_carts_user_id_idx;

-- Cart Items Table
DROP INDEX IF EXISTS cart_items_shopping_cart_id_idx;

-- Wishlists Table
DROP INDEX IF EXISTS wishlists_user_id_idx;

-- Wishlist Items Table
DROP INDEX IF EXISTS wishlist_items_wishlist_id_idx;

-- User Addresses Table
DROP INDEX IF EXISTS user_addresses_user_id_idx;

-- User Roles Table
DROP INDEX IF EXISTS user_roles_user_id_idx;

-- Refresh Tokens Table
DROP INDEX IF EXISTS refresh_tokens_user_id_idx;
DROP INDEX IF EXISTS refresh_tokens_expires_at_idx;

-- Shipment Table
DROP INDEX IF EXISTS shipment_order_id_idx;

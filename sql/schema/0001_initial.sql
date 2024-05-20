-- +goose Up

-- Users Table
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                       username VARCHAR(255) UNIQUE NOT NULL,
                       email VARCHAR(255) UNIQUE NOT NULL,
                       hashed_password VARCHAR(255) NOT NULL,
                       first_name VARCHAR(255),
                       last_name VARCHAR(255),
                       phone_number VARCHAR(20),
                       profile_image_url TEXT,
                       date_of_birth DATE,
                       is_active BOOLEAN NOT NULL DEFAULT TRUE,
                       created_at TIMESTAMPTZ DEFAULT NOW(),
                       updated_at TIMESTAMPTZ DEFAULT NOW(),
                       last_login TIMESTAMPTZ
);

-- Refresh Tokens Table
CREATE TABLE refresh_tokens (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                                token TEXT NOT NULL,
                                created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                                expires_at TIMESTAMPTZ NOT NULL,
                                revoked_at TIMESTAMPTZ,
                                UNIQUE (user_id, token)
);

-- User Addresses Table
CREATE TABLE user_addresses (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                address JSONB NOT NULL,
                                is_default BOOLEAN DEFAULT FALSE,
                                created_at TIMESTAMPTZ DEFAULT NOW(),
                                updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- User Roles Table
CREATE TABLE user_roles (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                            role VARCHAR(50) NOT NULL,
                            created_at TIMESTAMPTZ DEFAULT NOW(),
                            updated_at TIMESTAMPTZ DEFAULT NOW(),
                            UNIQUE (user_id, role)
);

-- Categories Table
CREATE TABLE categories (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            name VARCHAR(255) NOT NULL,
                            parent_id UUID REFERENCES categories(id) ON DELETE SET NULL,
                            created_at TIMESTAMPTZ DEFAULT NOW(),
                            updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Products Table
CREATE TABLE products (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          name VARCHAR(255) NOT NULL,
                          description TEXT,
                          price DECIMAL(10, 2) NOT NULL,
                          stock INT DEFAULT 0,
                          category_id UUID REFERENCES categories(id) ON DELETE SET NULL,
                          created_at TIMESTAMPTZ DEFAULT NOW(),
                          updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Variants Table
CREATE TABLE product_variants (
                                  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                  product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                  variant_name VARCHAR(255) NOT NULL,
                                  variant_value VARCHAR(255) NOT NULL,
                                  additional_price DECIMAL(10, 2) DEFAULT 0,
                                  created_at TIMESTAMPTZ DEFAULT NOW(),
                                  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Images Table
CREATE TABLE product_images (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                image_url TEXT NOT NULL,
                                created_at TIMESTAMPTZ DEFAULT NOW(),
                                updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Specifications Table
CREATE TABLE product_specifications (
                                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                        product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                        spec_name VARCHAR(255) NOT NULL,
                                        spec_value TEXT NOT NULL,
                                        created_at TIMESTAMPTZ DEFAULT NOW(),
                                        updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Orders Table
CREATE TABLE orders (
                        id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                        user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                        status VARCHAR(50) NOT NULL, -- pending, processing, shipped, delivered, cancelled
                        total DECIMAL(10, 2) NOT NULL,
                        created_at TIMESTAMPTZ DEFAULT NOW(),
                        updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Items Table
CREATE TABLE order_items (
                             id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                             order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
                             product_id UUID REFERENCES products(id) ON DELETE SET NULL,
                             product_variant_id UUID REFERENCES product_variants(id) ON DELETE SET NULL,
                             quantity INT NOT NULL,
                             price DECIMAL(10, 2) NOT NULL,
                             created_at TIMESTAMPTZ DEFAULT NOW(),
                             updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Status History Table
CREATE TABLE order_status_history (
                                      id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                      order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
                                      status VARCHAR(50) NOT NULL,
                                      created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Payments Table
CREATE TABLE order_payments (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
                                payment_id VARCHAR(255) NOT NULL,
                                status VARCHAR(50) NOT NULL, -- pending, paid, failed
                                method VARCHAR(50) NOT NULL, -- cash, credit_card, debit_card, paypal
                                amount DECIMAL(10, 2) NOT NULL,
                                created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Order Shipments Table
CREATE TABLE order_shipments (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
                                 tracking_id VARCHAR(255) NOT NULL,
                                 status VARCHAR(50) NOT NULL, -- pending, shipped, delivered
                                 created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Shopping Carts Table
CREATE TABLE shopping_carts (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                session_id UUID DEFAULT gen_random_uuid(),
                                total_items INT NOT NULL DEFAULT 0,
                                total_price DECIMAL(10, 2) NOT NULL DEFAULT 0.0,
                                created_at TIMESTAMPTZ DEFAULT NOW(),
                                updated_at TIMESTAMPTZ DEFAULT NOW(),
                                UNIQUE (user_id, session_id)
);

-- Cart Items Table
CREATE TABLE cart_items (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            shopping_cart_id UUID REFERENCES shopping_carts(id) ON DELETE CASCADE,
                            product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                            quantity INT NOT NULL,
                            created_at TIMESTAMPTZ DEFAULT NOW(),
                            updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Wishlists Table
CREATE TABLE wishlists (
                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                           user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                           name VARCHAR(50) NOT NULL,
                           created_at TIMESTAMPTZ DEFAULT NOW(),
                           updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Wishlist Items Table
CREATE TABLE wishlist_items (
                                id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                wishlist_id UUID REFERENCES wishlists(id) ON DELETE CASCADE,
                                product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                created_at TIMESTAMPTZ DEFAULT NOW(),
                                updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Product Reviews Table
CREATE TABLE product_reviews (
                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                                 product_id UUID REFERENCES products(id) ON DELETE CASCADE,
                                 user_id UUID REFERENCES users(id) ON DELETE CASCADE,
                                 rating INT CHECK (rating >= 1 AND rating <= 5),
                                 comment TEXT,
                                 created_at TIMESTAMPTZ DEFAULT NOW(),
                                 updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Shipment Table
CREATE TABLE shipment (
                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                          order_id UUID REFERENCES orders(id) ON DELETE CASCADE,
                          shipment_status VARCHAR(50) NOT NULL,
                          tracking_number VARCHAR(50),
                          shipped_date TIMESTAMPTZ,
                          delivery_date TIMESTAMPTZ,
                          created_at TIMESTAMPTZ DEFAULT NOW(),
                          updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- +goose Down
DROP TABLE IF EXISTS shipment;
DROP TABLE IF EXISTS product_reviews;
DROP TABLE IF EXISTS wishlist_items;
DROP TABLE IF EXISTS wishlists;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS shopping_carts;
DROP TABLE IF EXISTS order_shipments;
DROP TABLE IF EXISTS order_payments;
DROP TABLE IF EXISTS order_status_history;
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS product_specifications;
DROP TABLE IF EXISTS product_images;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS user_addresses;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
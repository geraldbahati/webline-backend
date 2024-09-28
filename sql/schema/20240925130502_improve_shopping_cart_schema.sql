-- +goose Up
-- Migration: 0001_improve_shopping_cart_schema.sql
-- Description: Improve shopping_carts and cart_items schemas for better performance and maintainability

BEGIN;

-- =========================================
-- Alter shopping_carts Table
-- =========================================

-- 1. Make user_id and session_id nullable
ALTER TABLE public.shopping_carts
    ALTER COLUMN user_id DROP NOT NULL,
    ALTER COLUMN session_id DROP NOT NULL;

-- 2. Remove unique constraint on (user_id, session_id) if it exists
ALTER TABLE public.shopping_carts
    DROP CONSTRAINT IF EXISTS shopping_carts_user_id_session_id_key;

-- 3. Rename constraints individually for clarity
ALTER TABLE public.shopping_carts
    RENAME CONSTRAINT check_total_items TO chk_total_items_non_negative;

ALTER TABLE public.shopping_carts
    RENAME CONSTRAINT check_total_price TO chk_total_price_non_negative;

ALTER TABLE public.shopping_carts
    RENAME CONSTRAINT shopping_carts_user_id_fkey TO fk_shopping_carts_user_id;

-- 4. Add separate indexes on user_id and session_id for performance
CREATE INDEX IF NOT EXISTS idx_shopping_carts_user_id ON public.shopping_carts(user_id);
CREATE INDEX IF NOT EXISTS idx_shopping_carts_session_id ON public.shopping_carts(session_id);

-- =========================================
-- Alter cart_items Table
-- =========================================

-- 1. Ensure shopping_cart_id and product_id are NOT NULL
ALTER TABLE public.cart_items
    ALTER COLUMN shopping_cart_id SET NOT NULL,
    ALTER COLUMN product_id SET NOT NULL;

-- 2. Rename constraints individually for clarity
ALTER TABLE public.cart_items
    RENAME CONSTRAINT cart_items_shopping_cart_id_fkey TO fk_cart_items_shopping_cart_id;

ALTER TABLE public.cart_items
    RENAME CONSTRAINT cart_items_product_id_fkey TO fk_cart_items_product_id;

ALTER TABLE public.cart_items
    RENAME CONSTRAINT unique_shopping_cart_product TO unique_cart_item;

-- 3. Add index on product_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_cart_items_product_id ON public.cart_items(product_id);

-- 4. Drop the disabled trigger if it's no longer needed
DROP TRIGGER IF EXISTS cart_items_before_insert ON public.cart_items;

-- =========================================
-- Optional: Create sessions Table
-- =========================================
-- Uncomment the following block if you decide to implement a sessions table for more detailed session management.

-- CREATE TABLE public.sessions (
--     id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
--     session_id UUID NOT NULL UNIQUE,
--     user_id UUID NULL,
--     created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
--     expires_at TIMESTAMPTZ NOT NULL,
--     CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
-- );
    
-- -- Indexes for sessions table
-- CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON public.sessions(user_id);

COMMIT;


-- +goose Down
-- Migration: 0001_improve_shopping_cart_schema.sql
-- Description: Revert improvements to shopping_carts and cart_items schemas

BEGIN;

-- =========================================
-- Revert Alterations on shopping_carts Table
-- =========================================

-- 1. Revert user_id and session_id to NOT NULL
ALTER TABLE public.shopping_carts
    ALTER COLUMN user_id SET NOT NULL,
    ALTER COLUMN session_id SET NOT NULL;

-- 2. Re-add unique constraint on (user_id, session_id)
ALTER TABLE public.shopping_carts
    ADD CONSTRAINT shopping_carts_user_id_session_id_key UNIQUE (user_id, session_id);

-- 3. Rename constraints back to original names
ALTER TABLE public.shopping_carts
    RENAME CONSTRAINT chk_total_items_non_negative TO check_total_items;

ALTER TABLE public.shopping_carts
    RENAME CONSTRAINT chk_total_price_non_negative TO check_total_price;

ALTER TABLE public.shopping_carts
    RENAME CONSTRAINT fk_shopping_carts_user_id TO shopping_carts_user_id_fkey;

-- 4. Drop the newly added indexes if they exist
DROP INDEX IF EXISTS idx_shopping_carts_user_id;
DROP INDEX IF EXISTS idx_shopping_carts_session_id;

-- =========================================
-- Revert Alterations on cart_items Table
-- =========================================

-- 1. Revert shopping_cart_id and product_id to nullable if they were previously nullable
-- Note: Only revert if they were previously nullable
-- ALTER TABLE public.cart_items
--     ALTER COLUMN shopping_cart_id DROP NOT NULL,
--     ALTER COLUMN product_id DROP NOT NULL;

-- 2. Rename constraints back to original names
ALTER TABLE public.cart_items
    RENAME CONSTRAINT fk_cart_items_shopping_cart_id TO cart_items_shopping_cart_id_fkey;

ALTER TABLE public.cart_items
    RENAME CONSTRAINT fk_cart_items_product_id TO cart_items_product_id_fkey;

ALTER TABLE public.cart_items
    RENAME CONSTRAINT unique_cart_item TO unique_shopping_cart_product;

-- 3. Drop the index on product_id if it exists
DROP INDEX IF EXISTS idx_cart_items_product_id;

-- 4. Recreate the previously dropped trigger if needed
-- Note: Only recreate if it was originally enabled and required
-- CREATE TRIGGER cart_items_before_insert
-- BEFORE INSERT ON public.cart_items
-- FOR EACH ROW
-- EXECUTE FUNCTION set_cart_item_price();

COMMIT;

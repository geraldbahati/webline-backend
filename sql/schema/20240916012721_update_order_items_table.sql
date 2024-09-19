-- +goose Up

BEGIN;

-- ========================================
-- Updates to the orders table
-- ========================================

-- Add new columns to the orders table
ALTER TABLE public.orders
ADD COLUMN IF NOT EXISTS currency_code CHAR(3) NOT NULL DEFAULT 'USD',
ADD COLUMN IF NOT EXISTS subtotal NUMERIC(12, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS tax_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS shipping_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS discount_amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS grand_total NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- Modify existing columns in orders table
ALTER TABLE public.orders
ALTER COLUMN status TYPE VARCHAR(20),
ALTER COLUMN status SET DEFAULT 'pending',
ALTER COLUMN payment_status TYPE VARCHAR(20),
ALTER COLUMN payment_status SET DEFAULT 'pending';

-- Update existing constraints on orders table
-- Drop old constraints if they exist
ALTER TABLE public.orders
DROP CONSTRAINT IF EXISTS check_status,
DROP CONSTRAINT IF EXISTS check_payment_status,
DROP CONSTRAINT IF EXISTS check_grand_total;

-- Add new constraints
ALTER TABLE public.orders
ADD CONSTRAINT check_status CHECK (status IN ('pending', 'processing', 'shipped', 'delivered', 'cancelled')),
ADD CONSTRAINT check_payment_status CHECK (payment_status IN ('pending', 'paid', 'failed', 'refunded')),
ADD CONSTRAINT check_grand_total CHECK (grand_total = subtotal + tax_amount + shipping_amount - discount_amount),
ADD CONSTRAINT check_user_or_guest CHECK (user_id IS NOT NULL OR guest_checkout_id IS NOT NULL);

-- Update foreign key constraints on orders table
ALTER TABLE public.orders
DROP CONSTRAINT IF EXISTS orders_user_id_fkey,
ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE public.orders
DROP CONSTRAINT IF EXISTS orders_guest_checkout_id_fkey,
ADD CONSTRAINT orders_guest_checkout_id_fkey FOREIGN KEY (guest_checkout_id) REFERENCES guest_checkouts(id) ON DELETE SET NULL;

-- Remove duplicate indexes from orders table
DROP INDEX IF EXISTS idx_order_status;
DROP INDEX IF EXISTS idx_order_user_id;
DROP INDEX IF EXISTS orders_status_idx;
DROP INDEX IF EXISTS orders_user_id_idx;
DROP INDEX IF EXISTS orders_user_status_idx;

-- Recreate necessary indexes for orders table
CREATE INDEX IF NOT EXISTS idx_orders_user_id ON public.orders(user_id);
CREATE INDEX IF NOT EXISTS idx_orders_status ON public.orders(status);
CREATE INDEX IF NOT EXISTS idx_orders_user_status ON public.orders(user_id, status);

-- ========================================
-- Updates to the order_items table
-- ========================================

-- Add new columns to order_items with temporary defaults
ALTER TABLE public.order_items
ADD COLUMN IF NOT EXISTS product_name VARCHAR(255) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS product_sku VARCHAR(100) NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS unit_price NUMERIC(12, 2) NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS total_price NUMERIC(12, 2) NOT NULL DEFAULT 0;

-- Update existing data in order_items
-- Use LEFT JOIN to handle cases where product_id is NULL
UPDATE public.order_items oi
SET product_name = COALESCE(p.name, ''),
    product_sku = COALESCE(p.part_number, ''),
    unit_price = CASE WHEN oi.quantity > 0 THEN oi.price / oi.quantity ELSE 0 END,
    total_price = oi.price
FROM public.products p
WHERE oi.product_id = p.id;

-- Remove temporary defaults from order_items
ALTER TABLE public.order_items
ALTER COLUMN product_name DROP DEFAULT,
ALTER COLUMN product_sku DROP DEFAULT,
ALTER COLUMN unit_price DROP DEFAULT,
ALTER COLUMN total_price DROP DEFAULT;

-- Rename and drop old price column in order_items
ALTER TABLE public.order_items
RENAME COLUMN price TO old_price;

-- Add constraint to ensure total_price equals quantity * unit_price
ALTER TABLE public.order_items
ADD CONSTRAINT check_total_price CHECK (total_price = quantity * unit_price);

-- Remove old price column from order_items
ALTER TABLE public.order_items
DROP COLUMN old_price;

-- Remove duplicate indexes from order_items
DROP INDEX IF EXISTS idx_order_item_order_id;
DROP INDEX IF EXISTS order_items_order_id_idx;

-- Recreate necessary index for order_items
CREATE INDEX IF NOT EXISTS idx_order_items_order_id ON public.order_items(order_id);

COMMIT;

-- +goose Down

BEGIN;

-- ========================================
-- Revert changes to the order_items table
-- ========================================

-- Add old price column back to order_items
ALTER TABLE public.order_items
ADD COLUMN price NUMERIC(10, 2) NOT NULL DEFAULT 0;

-- Restore price data in order_items
UPDATE public.order_items
SET price = total_price;

-- Remove new columns from order_items
ALTER TABLE public.order_items
DROP COLUMN IF EXISTS product_name,
DROP COLUMN IF EXISTS product_sku,
DROP COLUMN IF EXISTS unit_price,
DROP COLUMN IF EXISTS total_price;

-- Remove constraints from order_items
ALTER TABLE public.order_items
DROP CONSTRAINT IF EXISTS check_total_price;

-- Recreate original indexes for order_items
CREATE INDEX IF NOT EXISTS idx_order_item_order_id ON public.order_items(order_id);
CREATE INDEX IF NOT EXISTS order_items_order_id_idx ON public.order_items(order_id);

-- ========================================
-- Revert changes to the orders table
-- ========================================

-- Remove added columns from orders table
ALTER TABLE public.orders
DROP COLUMN IF EXISTS currency_code,
DROP COLUMN IF EXISTS subtotal,
DROP COLUMN IF EXISTS tax_amount,
DROP COLUMN IF EXISTS shipping_amount,
DROP COLUMN IF EXISTS discount_amount,
DROP COLUMN IF EXISTS grand_total;

-- Revert modifications to existing columns in orders table
ALTER TABLE public.orders
ALTER COLUMN status TYPE VARCHAR(50),
ALTER COLUMN status DROP DEFAULT,
ALTER COLUMN payment_status TYPE VARCHAR(50),
ALTER COLUMN payment_status SET DEFAULT 'pending';

-- Remove added constraints from orders table
ALTER TABLE public.orders
DROP CONSTRAINT IF EXISTS check_status,
DROP CONSTRAINT IF EXISTS check_payment_status,
DROP CONSTRAINT IF EXISTS check_grand_total,
DROP CONSTRAINT IF EXISTS check_user_or_guest;

-- Restore original foreign key constraints in orders table
ALTER TABLE public.orders
DROP CONSTRAINT IF EXISTS orders_user_id_fkey,
ADD CONSTRAINT orders_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE public.orders
DROP CONSTRAINT IF EXISTS orders_guest_checkout_id_fkey,
ADD CONSTRAINT orders_guest_checkout_id_fkey FOREIGN KEY (guest_checkout_id) REFERENCES guest_checkouts(id) ON DELETE CASCADE;

-- Recreate original indexes for orders table
CREATE INDEX IF NOT EXISTS idx_order_status ON public.orders(status);
CREATE INDEX IF NOT EXISTS idx_order_user_id ON public.orders(user_id);
CREATE INDEX IF NOT EXISTS orders_user_status_idx ON public.orders(user_id, status);

COMMIT;

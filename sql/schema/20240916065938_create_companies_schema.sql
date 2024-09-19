-- +goose Up

BEGIN;

-- ========================================
-- 1. Create the companies table
-- ========================================

CREATE TABLE IF NOT EXISTS public.companies (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    kra_pin VARCHAR(20) NOT NULL UNIQUE,
    address VARCHAR(255),
    phone_number VARCHAR(20),
    email VARCHAR(100),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_companies_kra_pin ON public.companies(kra_pin);
CREATE INDEX IF NOT EXISTS idx_companies_name ON public.companies(name);

-- ========================================
-- 2. Add company_id to the users table
-- ========================================

ALTER TABLE public.users
ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES public.companies(id) ON DELETE SET NULL;

-- Create index on company_id in users table
CREATE INDEX IF NOT EXISTS idx_users_company_id ON public.users(company_id);

-- ========================================
-- 3. Add company-related columns to the orders table
-- ========================================

ALTER TABLE public.orders
ADD COLUMN IF NOT EXISTS company_id UUID REFERENCES public.companies(id),
ADD COLUMN IF NOT EXISTS company_name VARCHAR(255),
ADD COLUMN IF NOT EXISTS kra_pin VARCHAR(20);

-- Create index on company_id in orders table
CREATE INDEX IF NOT EXISTS idx_orders_company_id ON public.orders(company_id);

-- ========================================
-- 4. Update existing orders with company information (if applicable)
-- ========================================

-- (Optional) If you have existing data and want to backfill company information, you can add SQL statements here.

-- ========================================
-- 5. Add triggers to update timestamps
-- ========================================

-- Create or replace the update_timestamp function if it doesn't exist
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_timestamp() RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Add trigger to companies table
CREATE TRIGGER update_companies_timestamp BEFORE UPDATE ON public.companies
FOR EACH ROW EXECUTE FUNCTION update_timestamp();

COMMIT;

-- +goose Down

BEGIN;

-- ========================================
-- 1. Remove company-related columns from orders table
-- ========================================

ALTER TABLE public.orders
DROP COLUMN IF EXISTS company_id,
DROP COLUMN IF EXISTS company_name,
DROP COLUMN IF EXISTS kra_pin;

-- Drop index on company_id in orders table
DROP INDEX IF EXISTS idx_orders_company_id;

-- ========================================
-- 2. Remove company_id from users table
-- ========================================

ALTER TABLE public.users
DROP COLUMN IF EXISTS company_id;

-- Drop index on company_id in users table
DROP INDEX IF EXISTS idx_users_company_id;

-- ========================================
-- 3. Drop the companies table
-- ========================================

DROP TABLE IF EXISTS public.companies;

-- ========================================
-- 4. Remove triggers and functions
-- ========================================

DROP TRIGGER IF EXISTS update_companies_timestamp ON public.companies;
DROP FUNCTION IF EXISTS update_timestamp();

-- ========================================
-- 5. Drop indexes on companies table
-- ========================================

DROP INDEX IF EXISTS idx_companies_kra_pin;
DROP INDEX IF EXISTS idx_companies_name;

COMMIT;

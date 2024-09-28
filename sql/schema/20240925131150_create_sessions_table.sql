-- +goose Up
-- Migration: 0002_create_sessions_table.sql
-- Description: Create sessions table and update shopping_carts to reference sessions

BEGIN;

-- 1. Create sessions table
CREATE TABLE public.sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL UNIQUE,
    user_id UUID NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_activity TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT fk_sessions_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- 2. Create indexes on sessions table
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON public.sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON public.sessions(expires_at);

-- 3. Add session_reference to shopping_carts
ALTER TABLE public.shopping_carts
    ADD COLUMN session_reference UUID NULL;

-- 4. Add foreign key constraint
ALTER TABLE public.shopping_carts
    ADD CONSTRAINT fk_shopping_carts_session_reference FOREIGN KEY (session_reference) REFERENCES sessions(session_id) ON DELETE SET NULL;

-- 5. Optionally, migrate existing session_id data to sessions table
-- This step requires data migration logic, possibly via a script or a more complex SQL migration.

-- 6. Drop the session_id column from shopping_carts if it's no longer needed
ALTER TABLE public.shopping_carts
    DROP COLUMN IF EXISTS session_id;

COMMIT;


-- +goose Down
-- Migration: 0002_create_sessions_table.sql
-- Description: Revert sessions table creation and update shopping_carts

BEGIN;

-- 1. Add session_id back to shopping_carts
ALTER TABLE public.shopping_carts
    ADD COLUMN session_id UUID DEFAULT gen_random_uuid();

-- 2. Remove foreign key constraint and session_reference column
ALTER TABLE public.shopping_carts
    DROP CONSTRAINT IF EXISTS fk_shopping_carts_session_reference,
    DROP COLUMN IF EXISTS session_reference;

-- 3. Drop sessions table
DROP TABLE IF EXISTS public.sessions CASCADE;

COMMIT;

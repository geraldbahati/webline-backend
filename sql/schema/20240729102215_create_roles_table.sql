-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Create Roles Table
CREATE TABLE public.roles (
                              id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
                              role_name character varying(50) NOT NULL,
                              description text,
                              created_at timestamp with time zone DEFAULT now(),
                              updated_at timestamp with time zone DEFAULT now(),
                              CONSTRAINT unique_role_name UNIQUE (role_name)
);

-- Update User Roles Table
ALTER TABLE public.user_roles
    ADD COLUMN role_id uuid REFERENCES roles(id) ON DELETE CASCADE,
    DROP CONSTRAINT user_roles_user_id_role_key;

ALTER TABLE public.user_roles
    ADD CONSTRAINT user_roles_user_id_role_id_key UNIQUE (user_id, role_id);

-- Remove role column from user_roles
ALTER TABLE public.user_roles
    DROP COLUMN role;

-- Indexes for roles table
CREATE INDEX roles_created_at_idx ON public.roles (created_at);

-- Triggers to update updated_at on modification
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ language 'plpgsql';
-- +goose StatementEnd

CREATE TRIGGER update_roles_updated_at BEFORE UPDATE
    ON public.roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_user_roles_updated_at BEFORE UPDATE
    ON public.user_roles FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back.

-- Drop Triggers
DROP TRIGGER IF EXISTS update_roles_updated_at ON public.roles;
DROP TRIGGER IF EXISTS update_user_roles_updated_at ON public.user_roles;

-- Drop Function
DROP FUNCTION IF EXISTS update_updated_at_column;

-- Revert User Roles Table changes
ALTER TABLE public.user_roles
    ADD COLUMN role character varying(50) NOT NULL default 'user',
    DROP CONSTRAINT user_roles_user_id_role_id_key;

ALTER TABLE public.user_roles
    ADD CONSTRAINT user_roles_user_id_role_key UNIQUE (user_id, role);

ALTER TABLE public.user_roles
    DROP COLUMN role_id;

-- Drop Roles Table
DROP TABLE IF EXISTS public.roles;

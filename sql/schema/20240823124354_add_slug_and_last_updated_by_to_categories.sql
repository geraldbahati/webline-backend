-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Add slug and last_updated_by columns
ALTER TABLE categories ADD COLUMN slug VARCHAR(255);
ALTER TABLE categories ADD COLUMN last_updated_by UUID REFERENCES users(id);

-- Function to generate slugs based on the ltree path
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION generate_slug_from_hierarchy(category_id UUID) RETURNS VARCHAR AS $$
DECLARE
    path ltree;
    slug VARCHAR;
BEGIN
    -- Generate the ltree path
    path := generate_ltree_path(category_id);

    -- Convert the ltree path to a slug by replacing dots with hyphens
    slug := regexp_replace(path::text, '\.', '-', 'g');

    RETURN slug;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Temporarily disable the trigger
ALTER TABLE categories DISABLE TRIGGER trg_update_category_path;

-- Update slugs for all existing categories and handle duplicates
-- +goose StatementBegin
DO $$
    DECLARE
        rec RECORD;
        new_slug VARCHAR;
        counter INTEGER;
    BEGIN
        FOR rec IN SELECT id FROM categories WHERE slug IS NULL LOOP
                BEGIN
                    -- Attempt to generate the slug
                    new_slug := generate_slug_from_hierarchy(rec.id);

                    -- Handle duplicates by appending a counter
                    counter := 1;
                    WHILE EXISTS (SELECT 1 FROM categories WHERE slug = new_slug) LOOP
                            new_slug := generate_slug_from_hierarchy(rec.id) || '-' || counter::text;
                            counter := counter + 1;
                        END LOOP;

                    -- Update the category with the new unique slug
                    UPDATE categories
                    SET slug = COALESCE(new_slug, 'unknown-slug')
                    WHERE id = rec.id;
                EXCEPTION WHEN OTHERS THEN
                    -- Log the failure
                    RAISE NOTICE 'Failed to update slug for category with ID: %, Error: %', rec.id, SQLERRM;
                    -- Set the slug to a default value in case of failure
                    UPDATE categories
                    SET slug = 'unknown-slug'
                    WHERE id = rec.id;
                END;
            END LOOP;

        -- Final safeguard: ensure all slugs are non-null
        UPDATE categories
        SET slug = 'unknown-slug'
        WHERE slug IS NULL;
    END $$;
-- +goose StatementEnd

-- Re-enable the trigger
ALTER TABLE categories ENABLE TRIGGER trg_update_category_path;

-- Ensure slug is unique and not null
ALTER TABLE categories ADD CONSTRAINT categories_slug_unique UNIQUE (slug);
ALTER TABLE categories ALTER COLUMN slug SET NOT NULL;

-- Trigger to update slug and last_updated_by before insert or update
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_category_slug() RETURNS TRIGGER AS $$
BEGIN
    -- Generate the ltree path based on the parent categories
    IF NEW.parent_id IS NULL THEN
        -- Root category
        NEW.path := COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'root')::ltree;
    ELSE
        -- Get the path of the parent category
        SELECT path
        INTO NEW.path
        FROM categories
        WHERE id = NEW.parent_id;

        -- Handle the case where the parent path is NULL
        IF NEW.path IS NULL THEN
            NEW.path := 'root'::ltree;
        END IF;

        -- Append the current category's name to the parent's path
        NEW.path := NEW.path || COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'child')::ltree;
    END IF;

    -- Convert the ltree path to a slug by replacing dots with hyphens
    NEW.slug := COALESCE(regexp_replace(NEW.path::text, '\.', '-', 'g'), 'unknown-slug');

    -- Handle slug duplicates by appending a counter
    DECLARE
        counter INTEGER := 1;
        base_slug VARCHAR := NEW.slug;
    BEGIN
        WHILE EXISTS (SELECT 1 FROM categories WHERE slug = NEW.slug AND id <> NEW.id) LOOP
                NEW.slug := base_slug || '-' || counter::text;
                counter := counter + 1;
            END LOOP;
    END;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_update_category_slug
    BEFORE INSERT OR UPDATE ON categories
    FOR EACH ROW
EXECUTE FUNCTION update_category_slug();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back.

-- Drop the trigger and functions
DROP TRIGGER trg_update_category_slug ON categories;
DROP FUNCTION update_category_slug;
DROP FUNCTION generate_slug_from_hierarchy;

-- Remove the slug and last_updated_by columns
ALTER TABLE categories DROP COLUMN slug;
ALTER TABLE categories DROP COLUMN last_updated_by;

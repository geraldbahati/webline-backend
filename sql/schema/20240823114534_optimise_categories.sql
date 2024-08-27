
-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Install the LTREE extension
CREATE EXTENSION IF NOT EXISTS ltree;

-- Add new columns to the categories table
ALTER TABLE categories ADD COLUMN description TEXT;
ALTER TABLE categories ADD COLUMN meta_title VARCHAR(255);
ALTER TABLE categories ADD COLUMN meta_description TEXT;
ALTER TABLE categories ADD COLUMN is_featured BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE categories ADD COLUMN level INTEGER NOT NULL DEFAULT 0;
ALTER TABLE categories ADD COLUMN path LTREE;
ALTER TABLE categories ADD COLUMN search_vector tsvector GENERATED ALWAYS AS (
    to_tsvector(
            'english',
            coalesce(name, '') || ' ' ||
            coalesce(description, '') || ' ' ||
            coalesce(meta_title, '') || ' ' ||
            coalesce(meta_description, '')
    )
    ) STORED;

-- Create indexes for path and search vector
CREATE INDEX idx_categories_path ON categories USING GIST(path);
CREATE INDEX idx_categories_search_vector ON categories USING GIN(search_vector);

-- Function to generate ltree path from sanitized category names
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION generate_ltree_path(category_id UUID) RETURNS ltree AS $$
DECLARE
    path ltree;
BEGIN
    SELECT
        string_agg(regexp_replace(lower(name), '[^a-z0-9]+', '_', 'g'), '.')::ltree
    INTO path
    FROM (
             WITH RECURSIVE parent_categories AS (
                 SELECT id, parent_id, name, 0 AS depth
                 FROM categories
                 WHERE id = category_id
                 UNION ALL
                 SELECT c.id, c.parent_id, c.name, pc.depth + 1
                 FROM categories c
                          INNER JOIN parent_categories pc ON pc.parent_id = c.id
             )
             SELECT name FROM parent_categories ORDER BY depth DESC
         ) AS paths;

    RETURN path;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Update paths and levels for all existing categories
-- +goose StatementBegin
DO $$
    DECLARE
        rec RECORD;
        new_path ltree;
    BEGIN
        FOR rec IN SELECT id FROM categories LOOP
                new_path := generate_ltree_path(rec.id);
                UPDATE categories
                SET path = new_path,
                    level = nlevel(new_path)
                WHERE id = rec.id;
            END LOOP;
    END $$;
-- +goose StatementEnd

-- Create a trigger for future updates and inserts
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_category_path() RETURNS TRIGGER AS $$
BEGIN
    IF NEW.parent_id IS NULL THEN
        -- Set path based on category name if no parent
        NEW.path := COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'root')::ltree;
    ELSE
        -- Fetch the parent path
        SELECT path INTO NEW.path
        FROM categories
        WHERE id = NEW.parent_id;

        -- Handle case where parent path is NULL
        IF NEW.path IS NULL THEN
            NEW.path := 'root'::ltree;
        END IF;

        -- Safely append current category name to the parent's path
        NEW.path := (NEW.path || COALESCE(NULLIF(regexp_replace(lower(NEW.name), '[^a-z0-9]+', '_', 'g'), ''), 'child'))::ltree;
    END IF;

    -- Update the level based on the new path
    NEW.level := nlevel(NEW.path);

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

CREATE TRIGGER trg_update_category_path
    BEFORE INSERT OR UPDATE ON categories
    FOR EACH ROW
EXECUTE FUNCTION update_category_path();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back.

-- Drop the trigger and function
DROP TRIGGER trg_update_category_path ON categories;
DROP FUNCTION update_category_path;
DROP FUNCTION generate_ltree_path;

-- Drop the indexes
DROP INDEX IF EXISTS idx_categories_path;
DROP INDEX IF EXISTS idx_categories_search_vector;

-- Remove the columns
ALTER TABLE categories DROP COLUMN search_vector;
ALTER TABLE categories DROP COLUMN path;
ALTER TABLE categories DROP COLUMN level;
ALTER TABLE categories DROP COLUMN is_featured;
ALTER TABLE categories DROP COLUMN meta_description;
ALTER TABLE categories DROP COLUMN meta_title;
ALTER TABLE categories DROP COLUMN description;

-- Remove the LTREE extension
DROP EXTENSION IF EXISTS ltree;

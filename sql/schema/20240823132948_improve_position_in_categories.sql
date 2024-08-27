-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Create a function to automatically set the position
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_category_position() RETURNS TRIGGER AS $$
BEGIN
    -- If no position is provided, find the next available position within the same parent
    IF NEW.position = -1 THEN
        NEW.position := COALESCE(
                (SELECT MAX(position) + 1 FROM categories WHERE parent_id = NEW.parent_id),
                1
                        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create a trigger to set the position before inserting a new category
CREATE TRIGGER trg_set_category_position
    BEFORE INSERT ON categories
    FOR EACH ROW EXECUTE FUNCTION set_category_position();

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back.

-- Drop the trigger and function
DROP TRIGGER IF EXISTS trg_set_category_position ON categories;
DROP FUNCTION IF EXISTS set_category_position;

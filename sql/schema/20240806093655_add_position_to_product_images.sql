-- +goose Up
-- Add the `position` column
ALTER TABLE product_images
    ADD COLUMN position integer;

-- Create the trigger function
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION set_default_position()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.position IS NULL THEN
        NEW.position := COALESCE(
                (SELECT MAX(position) + 1 FROM product_images WHERE product_id = NEW.product_id),
                1
                        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- Create the trigger
-- +goose StatementBegin
CREATE TRIGGER set_position_before_insert
    BEFORE INSERT ON product_images
    FOR EACH ROW
EXECUTE FUNCTION set_default_position();
-- +goose StatementEnd

-- +goose Down
-- Remove the trigger and trigger function
DROP TRIGGER IF EXISTS set_position_before_insert ON product_images;
DROP FUNCTION IF EXISTS set_default_position();

-- Remove the `position` column
ALTER TABLE product_images
    DROP COLUMN position;

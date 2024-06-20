-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION format_phone_number()
RETURNS TRIGGER AS $$
BEGIN
    -- For users table
    IF TG_TABLE_NAME = 'users' THEN
        NEW.phone_number = '254' || SUBSTRING(NEW.phone_number FROM 2);
    END IF;

    -- For guest_checkouts table
    IF TG_TABLE_NAME = 'guest_checkouts' THEN
        NEW.phone = '254' || SUBSTRING(NEW.phone FROM 2);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER format_phone_number_users
BEFORE INSERT OR UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION format_phone_number();

CREATE TRIGGER format_phone_number_guest_checkouts
BEFORE INSERT OR UPDATE ON guest_checkouts
FOR EACH ROW
EXECUTE FUNCTION format_phone_number();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS format_phone_number_users ON users;
DROP TRIGGER IF EXISTS format_phone_number_guest_checkouts ON guest_checkouts;
DROP FUNCTION IF EXISTS format_phone_number();
-- +goose StatementEnd

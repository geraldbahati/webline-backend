-- +goose Up
-- +goose StatementBegin
ALTER TABLE promotions
    ADD COLUMN slug VARCHAR(255) UNIQUE NOT NULL default '',
    ADD CONSTRAINT promotions_slug_unique UNIQUE(slug);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE FUNCTION generate_promotion_slug() RETURNS TRIGGER AS $$
DECLARE
    temp_slug VARCHAR(255);
    suffix INT := 1;
BEGIN
    -- Generate the initial slug
    temp_slug := lower(regexp_replace(NEW.title, '[^a-zA-Z0-9]+', '-', 'g'));

    -- Check for uniqueness and append a suffix if necessary
    WHILE EXISTS (SELECT 1 FROM promotions WHERE slug = temp_slug) LOOP
            temp_slug := lower(regexp_replace(NEW.title, '[^a-zA-Z0-9]+', '-', 'g')) || '-' || suffix;
            suffix := suffix + 1;
        END LOOP;

    NEW.slug := temp_slug;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER generate_slug_trigger
    BEFORE INSERT ON promotions
    FOR EACH ROW
EXECUTE FUNCTION generate_promotion_slug();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_promotions_slug ON promotions(slug);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS generate_slug_trigger ON promotions;
-- +goose StatementEnd

-- +goose StatementBegin
DROP FUNCTION IF EXISTS generate_promotion_slug();
-- +goose StatementEnd

-- +goose StatementBegin
DROP INDEX IF EXISTS idx_promotions_slug;
-- +goose StatementEnd

-- +goose StatementBegin
ALTER TABLE promotions DROP CONSTRAINT IF EXISTS promotions_slug_unique;
ALTER TABLE promotions DROP COLUMN IF EXISTS slug;
-- +goose StatementEnd


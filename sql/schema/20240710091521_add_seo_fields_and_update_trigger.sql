-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_search_keyword()
    RETURNS trigger AS $$
BEGIN
    NEW.search_keyword :=
            setweight(to_tsvector(coalesce(NEW.name, '')), 'A') ||
            setweight(to_tsvector(coalesce(NEW.description, '')), 'B') ||
            setweight(to_tsvector(coalesce(NEW.part_number, '')), 'A') ||
            setweight(to_tsvector(coalesce(NEW.meta_title, '')), 'C') ||
            setweight(to_tsvector(coalesce(NEW.meta_description, '')), 'D') ||
            setweight(to_tsvector(coalesce(NEW.meta_keywords, '')), 'D');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_search_keyword()
    RETURNS trigger AS $$
BEGIN
    NEW.search_keyword :=
            setweight(to_tsvector(coalesce(NEW.name, '')), 'A') ||
            setweight(to_tsvector(coalesce(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

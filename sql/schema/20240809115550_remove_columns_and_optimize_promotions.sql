-- +goose Up
-- Migration to remove columns and optimize promotions table

BEGIN;

ALTER TABLE promotions
    DROP COLUMN tagline,
    DROP COLUMN subtitle,
    DROP COLUMN main_title;

ALTER TABLE promotions
    ALTER COLUMN start_date SET NOT NULL,
    ALTER COLUMN end_date SET NOT NULL;

CREATE INDEX idx_promotions_start_date ON promotions(start_date);
CREATE INDEX idx_promotions_end_date ON promotions(end_date);

DROP INDEX IF EXISTS idx_promotions_main_title;
DROP INDEX IF EXISTS idx_promotions_title;

COMMIT;

-- +goose Down
-- Revert migration by restoring removed columns and indexes

BEGIN;

ALTER TABLE promotions
    ADD COLUMN tagline character varying(255),
    ADD COLUMN subtitle character varying(255) NOT NULL DEFAULT ''::character varying,
    ADD COLUMN main_title character varying(255) NOT NULL DEFAULT ''::character varying;

ALTER TABLE promotions
    ALTER COLUMN start_date DROP NOT NULL,
    ALTER COLUMN end_date DROP NOT NULL;

DROP INDEX IF EXISTS idx_promotions_start_date;
DROP INDEX IF EXISTS idx_promotions_end_date;

CREATE INDEX idx_promotions_main_title ON promotions(main_title);
CREATE INDEX idx_promotions_title ON promotions(title);

COMMIT;

-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

ALTER TABLE public.promotions
    ADD COLUMN tagline character varying(255),
    ADD COLUMN main_title character varying(255) NOT NULL default '',
    ADD COLUMN subtitle character varying(255) NOT NULL default '',
    ADD COLUMN start_date timestamp with time zone,
    ADD COLUMN end_date timestamp with time zone;

-- Add index for main_title
CREATE INDEX idx_promotions_main_title ON public.promotions (main_title);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back.

ALTER TABLE public.promotions
    DROP COLUMN IF EXISTS tagline,
    DROP COLUMN IF EXISTS main_title,
    DROP COLUMN IF EXISTS subtitle,
    DROP COLUMN IF EXISTS start_date,
    DROP COLUMN IF EXISTS end_date;

DROP INDEX IF EXISTS idx_promotions_main_title;

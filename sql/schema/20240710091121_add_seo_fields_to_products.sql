-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE public.products
    ADD COLUMN part_number character varying(100),
    ADD COLUMN meta_title character varying(255),
    ADD COLUMN meta_description text,
    ADD COLUMN meta_keywords character varying(255);

-- Create index for part_number
CREATE INDEX idx_products_part_number ON public.products(part_number);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

ALTER TABLE public.products
    DROP COLUMN part_number,
    DROP COLUMN meta_title,
    DROP COLUMN meta_description,
    DROP COLUMN meta_keywords;

-- Drop index for part_number
DROP INDEX IF EXISTS idx_products_part_number;

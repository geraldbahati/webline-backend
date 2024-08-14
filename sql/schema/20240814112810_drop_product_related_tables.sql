-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

BEGIN;

DROP TABLE IF EXISTS product_processors CASCADE;
DROP TABLE IF EXISTS product_sizes CASCADE;
DROP TABLE IF EXISTS product_storage_options CASCADE;
DROP TABLE IF EXISTS sizes CASCADE;
DROP TABLE IF EXISTS storage_options CASCADE;
DROP TABLE IF EXISTS colors CASCADE;
DROP TABLE IF EXISTS product_colors CASCADE;
DROP TABLE IF EXISTS processors CASCADE;

COMMIT;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back

BEGIN;

CREATE TABLE colors (
                        id          UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                        color_name  VARCHAR(50) NOT NULL UNIQUE,
                        created_at  TIMESTAMP WITH TIME ZONE DEFAULT now(),
                        updated_at  TIMESTAMP WITH TIME ZONE DEFAULT now(),
                        color_value VARCHAR(7)
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE TABLE processors (
                            id         UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                            name       VARCHAR(255) NOT NULL,
                            created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                            updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE INDEX idx_processors_name ON processors (name);

CREATE TABLE sizes (
                       id         UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                       size       VARCHAR(50) NOT NULL UNIQUE,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE TABLE storage_options (
                                 id         UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                 name       VARCHAR(255) NOT NULL,
                                 created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                 updated_at TIMESTAMP WITH TIME ZONE DEFAULT now()
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE INDEX idx_storage_options_name ON storage_options (name);

CREATE TABLE product_colors (
                                id         UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                product_id UUID REFERENCES products ON DELETE CASCADE,
                                created_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                updated_at TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                color_id   UUID CONSTRAINT fk_color_id REFERENCES colors ON DELETE RESTRICT
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE INDEX idx_product_colors_product_id ON product_colors (product_id);
CREATE INDEX idx_product_colors_color_id ON product_colors (color_id);

CREATE TABLE product_processors (
                                    id           UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                    product_id   UUID REFERENCES products ON DELETE CASCADE,
                                    processor_id UUID REFERENCES processors ON DELETE CASCADE,
                                    created_at   TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                    updated_at   TIMESTAMP WITH TIME ZONE DEFAULT now()
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE INDEX idx_product_processors_product_id ON product_processors (product_id);
CREATE INDEX idx_product_processors_processor_id ON product_processors (processor_id);

CREATE TABLE product_storage_options (
                                         id                UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                                         product_id        UUID REFERENCES products ON DELETE CASCADE,
                                         storage_option_id UUID REFERENCES storage_options ON DELETE CASCADE,
                                         additional_price  NUMERIC(10, 2) DEFAULT 0,
                                         created_at        TIMESTAMP WITH TIME ZONE DEFAULT now(),
                                         updated_at        TIMESTAMP WITH TIME ZONE DEFAULT now()
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE INDEX idx_product_storage_options_product_id ON product_storage_options (product_id);
CREATE INDEX idx_product_storage_options_storage_option_id ON product_storage_options (storage_option_id);

CREATE TABLE product_sizes (
                               id               UUID DEFAULT gen_random_uuid() PRIMARY KEY,
                               product_id       UUID REFERENCES products ON DELETE CASCADE,
                               additional_price NUMERIC(10, 2) DEFAULT 0,
                               created_at       TIMESTAMP WITH TIME ZONE DEFAULT now(),
                               updated_at       TIMESTAMP WITH TIME ZONE DEFAULT now(),
                               size_id          UUID CONSTRAINT fk_size_id REFERENCES sizes ON DELETE RESTRICT
) WITH (OIDS=FALSE) OWNER TO geraldbahati;

CREATE INDEX idx_product_sizes_product_id ON product_sizes (product_id);
CREATE INDEX idx_product_sizes_size_id ON product_sizes (size_id);

COMMIT;

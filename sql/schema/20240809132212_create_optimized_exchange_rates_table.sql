-- +goose Up
-- Create optimized exchange_rates table

CREATE TABLE exchange_rates (
                                id SERIAL PRIMARY KEY,
                                currency_code VARCHAR(3) NOT NULL,
                                rate_to_kes NUMERIC(10, 4) NOT NULL CHECK (rate_to_kes > 0),
                                valid_from TIMESTAMP WITH TIME ZONE NOT NULL,
                                valid_to TIMESTAMP WITH TIME ZONE,
                                CONSTRAINT unique_currency_valid_range UNIQUE (currency_code, valid_from, valid_to),
                                CONSTRAINT check_validity_range CHECK (valid_to IS NULL OR valid_to > valid_from)
);

-- Index for efficient lookups by currency code and validity period
CREATE INDEX idx_exchange_rates_validity ON exchange_rates (currency_code, valid_from DESC, valid_to DESC);

-- +goose Down
-- Drop the exchange_rates table

DROP INDEX IF EXISTS idx_exchange_rates_validity;
DROP TABLE IF EXISTS exchange_rates;

-- +goose Up
ALTER TABLE products
    ALTER COLUMN usd_price TYPE numeric(15,10),
    ALTER COLUMN price_per_unit TYPE numeric(15,10);

-- +goose Down
ALTER TABLE products
    ALTER COLUMN usd_price TYPE numeric(10,2),
    ALTER COLUMN price_per_unit TYPE numeric(10,2);

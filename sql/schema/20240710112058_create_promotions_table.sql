-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE promotions (
                            id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
                            title VARCHAR(255) NOT NULL,
                            description TEXT,
                            image_url VARCHAR(255),
                            created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                            updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE promotion_products (
                                    promotion_id UUID NOT NULL,
                                    product_id UUID NOT NULL,
                                    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                                    PRIMARY KEY (promotion_id, product_id),
                                    FOREIGN KEY (promotion_id) REFERENCES promotions(id) ON DELETE CASCADE,
                                    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE CASCADE
);

-- Add Indexes
CREATE INDEX idx_promotions_title ON promotions(title);
CREATE INDEX idx_promotion_products_promotion_id ON promotion_products(promotion_id);
CREATE INDEX idx_promotion_products_product_id ON promotion_products(product_id);

-- Triggers to update updated_at column
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION update_updated_at_column()
    RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trigger_update_promotions_updated_at
    BEFORE UPDATE ON promotions
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER trigger_update_promotion_products_updated_at
    BEFORE UPDATE ON promotion_products
    FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS promotion_products CASCADE;
DROP TABLE IF EXISTS promotions CASCADE;

DROP FUNCTION IF EXISTS update_updated_at_column CASCADE;

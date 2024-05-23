-- +goose Up
-- Create Product Interactions Table
CREATE TABLE IF NOT EXISTS product_interactions (
                                                    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID REFERENCES products(id) ON DELETE CASCADE,
    interaction_type VARCHAR(50) NOT NULL, -- e.g., 'view', 'purchase'
    user_id UUID REFERENCES users(id) ON DELETE SET NULL, -- Can be null if not logged in
    interaction_time TIMESTAMPTZ DEFAULT NOW()
    );

CREATE INDEX IF NOT EXISTS idx_product_interactions_product_id ON product_interactions(product_id);
CREATE INDEX IF NOT EXISTS idx_product_interactions_user_id ON product_interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_product_interactions_interaction_type ON product_interactions(interaction_type);
CREATE INDEX IF NOT EXISTS idx_product_interactions_interaction_time ON product_interactions(interaction_time);

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Drop Product Interactions Table
DROP TABLE IF EXISTS product_interactions;

-- Drop Indexes
DROP INDEX IF EXISTS idx_product_interactions_product_id;
DROP INDEX IF EXISTS idx_product_interactions_user_id;
DROP INDEX IF EXISTS idx_product_interactions_interaction_type;
DROP INDEX IF EXISTS idx_product_interactions_interaction_time;
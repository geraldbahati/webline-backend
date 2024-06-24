-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

ALTER TABLE product_interactions
ADD CONSTRAINT check_interaction_type CHECK (interaction_type IN ('view', 'click', 'purchase'));

CREATE INDEX IF NOT EXISTS idx_product_interactions_product_id ON product_interactions(product_id);
CREATE INDEX IF NOT EXISTS idx_product_interactions_user_id ON product_interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_product_interactions_interaction_time ON product_interactions(interaction_time);
CREATE INDEX IF NOT EXISTS idx_product_interactions_interaction_type ON product_interactions(interaction_type);


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

ALTER TABLE product_interactions
DROP CONSTRAINT IF EXISTS check_interaction_type;

DROP INDEX IF EXISTS idx_product_interactions_product_id;
DROP INDEX IF EXISTS idx_product_interactions_user_id;
DROP INDEX IF EXISTS idx_product_interactions_interaction_time;
DROP INDEX IF EXISTS idx_product_interactions_interaction_type;

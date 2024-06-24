-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
DROP VIEW IF EXISTS user_preferences;

CREATE MATERIALIZED VIEW user_preferences AS
SELECT user_id, product_id, COUNT(*) AS interaction_count
FROM product_interactions
GROUP BY user_id, product_id;

CREATE INDEX idx_user_preferences_user_id ON user_preferences(user_id);
CREATE INDEX idx_user_preferences_product_id ON user_preferences(product_id);
-- +goose StatementEnd

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS user_preferences;
-- +goose StatementEnd

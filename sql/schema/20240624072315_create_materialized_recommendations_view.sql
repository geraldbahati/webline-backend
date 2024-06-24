-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

-- +goose StatementBegin
DROP VIEW IF EXISTS recommendations;

CREATE MATERIALIZED VIEW recommendations AS
SELECT user_id, product_id, COUNT(*) AS score
FROM product_interactions
GROUP BY user_id, product_id;

CREATE INDEX idx_recommendations_user_id ON recommendations(user_id);
CREATE INDEX idx_recommendations_product_id ON recommendations(product_id);
-- +goose StatementEnd


-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS recommendations;
-- +goose StatementEnd

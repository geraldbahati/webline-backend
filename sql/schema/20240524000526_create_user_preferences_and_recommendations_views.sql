-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied.

-- Drop View for User Preferences if it exists
DROP VIEW IF EXISTS user_preferences;

-- Create View for User Preferences
CREATE VIEW user_preferences AS
SELECT
    user_id,
    product_id,
    COUNT(*) AS interaction_count
FROM
    product_interactions
GROUP BY
    user_id, product_id
ORDER BY
    user_id, interaction_count DESC;

-- Drop View for Recommendations if it exists
DROP VIEW IF EXISTS recommendations;

-- Create View for Recommendations
CREATE VIEW recommendations AS
SELECT
    up1.user_id,
    up2.product_id,
    SUM(up2.interaction_count) AS score
FROM
    user_preferences up1
        JOIN
    user_preferences up2 ON up1.product_id = up2.product_id AND up1.user_id <> up2.user_id
WHERE
    up2.product_id NOT IN (SELECT product_id FROM user_preferences WHERE user_id = up1.user_id)
GROUP BY
    up1.user_id, up2.product_id
ORDER BY
    up1.user_id, score DESC;

-- +goose Down
-- SQL in section 'Down' is executed when this migration is rolled back.

-- Drop View for Recommendations
DROP VIEW IF EXISTS recommendations;

-- Drop View for User Preferences
DROP VIEW IF EXISTS user_preferences;

-- +goose Up
-- SQL statements for the "up" migration.

BEGIN;

-- Create Materialized View for category hierarchy ordered by position
CREATE MATERIALIZED VIEW IF NOT EXISTS category_hierarchy_mv AS
WITH RECURSIVE category_hierarchy AS (
    SELECT
        id,
        name,
        parent_id,
        position
    FROM
        categories
    WHERE
        parent_id IS NULL

    UNION ALL

    SELECT
        c.id,
        c.name,
        c.parent_id,
        c.position
    FROM
        categories c
            JOIN category_hierarchy ch ON c.parent_id = ch.id
)
SELECT * FROM category_hierarchy
ORDER BY position;

-- Optionally create more views or optimized tables
CREATE MATERIALIZED VIEW rate_mv AS
SELECT COALESCE(
               (SELECT rate_to_kes
                FROM exchange_rates
                WHERE currency_code = 'USD'
                  AND (valid_to IS NULL OR valid_to >= NOW())
                  AND valid_from <= NOW()
                ORDER BY valid_from DESC
                LIMIT 1),
               135) AS rate_to_kes;


COMMIT;

-- +goose Down
-- SQL statements for the "down" migration.

BEGIN;

-- Drop Materialized View
DROP MATERIALIZED VIEW IF EXISTS category_hierarchy_mv;

-- Optionally drop other views or optimized tables

COMMIT;

-- +goose Up
-- +goose StatementBegin
WITH duplicates AS (
    SELECT user_id, COUNT(*) as count
    FROM shopping_carts
    WHERE user_id IS NOT NULL
    GROUP BY user_id
    HAVING COUNT(*) > 1
)
-- Delete all duplicates except the most recent one
DELETE FROM shopping_carts
WHERE id IN (
    SELECT sc.id
    FROM shopping_carts sc
    INNER JOIN (
        SELECT id, user_id, ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY updated_at DESC) as rn
        FROM shopping_carts
        WHERE user_id IS NOT NULL
    ) sub ON sc.id = sub.id
    WHERE sub.rn > 1
);
-- +goose StatementEnd

-- Add the unique constraint
ALTER TABLE shopping_carts
ADD CONSTRAINT unique_user_id UNIQUE (user_id);

-- +goose Down
-- +goose StatementBegin
-- Remove the unique constraint

ALTER TABLE shopping_carts
DROP CONSTRAINT unique_user_id;
-- +goose StatementEnd

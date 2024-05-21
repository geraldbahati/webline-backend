-- name: StoreRefreshToken :one
INSERT INTO refresh_tokens (id, user_id, token, created_at, expires_at)
    VALUES (gen_random_uuid(), $1, $2, NOW(), $3)
RETURNING id, user_id, token, created_at, expires_at;

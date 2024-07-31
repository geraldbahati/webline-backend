-- name: StorePasswordResetToken :exec
INSERT INTO password_reset_tokens (email, token, expires_at)
VALUES ($1, $2, $3);

-- name: GetPasswordResetToken :one
SELECT id, email, token, expires_at, created_at
FROM password_reset_tokens
WHERE email = $1;

-- name: DeletePasswordResetToken :exec
DELETE FROM password_reset_tokens
WHERE email = $1;

-- name: DeleteExpiredTResets :exec
DELETE FROM password_reset_tokens
WHERE expires_at < now();

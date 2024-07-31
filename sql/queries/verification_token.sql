-- name: CreateVerificationToken :one
INSERT INTO verification_tokens (
    email, token, expires_at
) VALUES (
             $1, $2, $3
         )
RETURNING id, email, token, expires_at, created_at;

-- name: GetVerificationTokenByToken :one
SELECT id, email, token, expires_at, created_at
FROM verification_tokens
WHERE token = $1;

-- name: DeleteVerificationTokensByEmail :exec
DELETE FROM verification_tokens
WHERE email = $1;

-- name: DeleteVerificationTokenByToken :exec
DELETE FROM verification_tokens
WHERE token = $1;

-- name: DeleteExpiredTokens :exec
DELETE FROM verification_tokens
WHERE expires_at < now();

-- name: GetVerificationTokenByEmail :one
SELECT id, email, token, expires_at, created_at
FROM verification_tokens
WHERE email = $1
ORDER BY created_at DESC
LIMIT 1;

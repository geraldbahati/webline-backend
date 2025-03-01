-- name: CreateSession :one
INSERT INTO sessions (
    user_id,
    session_id,
    csrf_token,
    expires_at
) VALUES (
    $1,
    $2,
    $3,
    $4
) RETURNING id, session_id, user_id, created_at, last_activity, expires_at, csrf_token;

-- name: GetSessionBySessionID :one
SELECT id, session_id, user_id, created_at, last_activity, expires_at, csrf_token
FROM sessions
WHERE session_id = $1;

-- name: GetSessionByUserID :many
SELECT id, session_id, user_id, created_at, last_activity, expires_at, csrf_token
FROM sessions
WHERE user_id = $1;

-- name: LinkSessionToUser :exec
UPDATE sessions
SET user_id = $2, last_activity = NOW()
WHERE session_id = $1;

-- name: DeleteSessionBySessionID :exec
DELETE FROM sessions
WHERE session_id = $1;

-- name: UpdateSessionLastActivity :exec
UPDATE sessions
SET last_activity = NOW()
WHERE session_id = $1;

-- name: UpdateSession :exec
UPDATE sessions
SET user_id = $2, last_activity = NOW(), expires_at = $3, csrf_token = $4
WHERE session_id = $1;

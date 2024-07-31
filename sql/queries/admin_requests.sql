-- name: CreateAdminRequest :one
INSERT INTO admin_requests ( user_id, reason)
VALUES ($1, $2)
RETURNING id;

-- name: GetPendingAdminRequests :many
SELECT
    ar.id,
    ar.user_id,
    ar.reason,
    ar.status,
    ar.created_at,
    ar.updated_at,
    u.email
FROM
    admin_requests ar
        JOIN
    users u ON ar.user_id = u.id
WHERE
    ar.status = 'PENDING'
ORDER BY
    ar.created_at;


-- name: GetAdminRequestByID :one
SELECT
    ar.id,
    ar.user_id,
    ar.reason,
    ar.status,
    ar.created_at,
    ar.updated_at,
    u.email
FROM
    admin_requests ar
        JOIN
    users u ON ar.user_id = u.id
WHERE
    ar.id = $1;


-- name: ApproveAdminRequest :exec
UPDATE admin_requests
SET status = 'APPROVED', updated_at = now()
WHERE id = $1 AND status = 'PENDING';

-- name: RejectAdminRequest :exec
UPDATE admin_requests
SET status = 'REJECTED', updated_at = now()
WHERE id = $1 AND status = 'PENDING';

-- name: GetAdminRequestsByUserID :many
SELECT
    ar.id,
    ar.user_id,
    ar.reason,
    ar.status,
    ar.created_at,
    ar.updated_at,
    u.email
FROM
    admin_requests ar
        JOIN
    users u ON ar.user_id = u.id
WHERE
    ar.user_id = $1
ORDER BY
    ar.created_at DESC;

-- name: StoreApprovalToken :exec
INSERT INTO admin_approval_tokens ( request_id, token, expires_at)
VALUES ($1, $2, $3);

-- name: GetApprovalToken :one
SELECT id, request_id, token, expires_at, created_at
FROM admin_approval_tokens
WHERE token = $1;

-- name: DeleteApprovalToken :exec
DELETE FROM admin_approval_tokens
WHERE token = $1;

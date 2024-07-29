-- name: AssignRoleToUser :one
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
RETURNING id, user_id, role_id, created_at, updated_at;

-- name: GetUserRoles :many
SELECT ur.id, ur.user_id, ur.role_id, r.role_name, ur.created_at, ur.updated_at
FROM user_roles ur
         JOIN roles r ON ur.role_id = r.id
WHERE ur.user_id = $1;

-- name: GetUsersByRole :many
SELECT u.id, u.email, u.first_name, u.last_name, u.phone_number, u.profile_image_url, u.date_of_birth, u.is_active, u.created_at, u.updated_at, u.last_login
FROM users u
         JOIN user_roles ur ON u.id = ur.user_id
WHERE ur.role_id = $1;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: RemoveAllRolesFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1;

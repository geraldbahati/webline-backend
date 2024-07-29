-- name: CreateRole :one
INSERT INTO roles (role_name, description)
VALUES ($1, $2)
RETURNING id, role_name, description, created_at, updated_at;

-- name: GetRoleByID :one
SELECT id, role_name, description, created_at, updated_at
FROM roles
WHERE id = $1;

-- name: GetRoleByName :one
SELECT id, role_name
FROM roles
WHERE role_name = $1;

-- name: GetAllRoles :many
SELECT id, role_name, description, created_at, updated_at
FROM roles
ORDER BY role_name;

-- name: UpdateRole :one
UPDATE roles
SET role_name = $2, description = $3, updated_at = now()
WHERE id = $1
RETURNING id, role_name, description, created_at, updated_at;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE id = $1;

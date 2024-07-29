-- name: CreateUser :one
INSERT INTO users ( email, hashed_password)
VALUES ( $1, $2) RETURNING id, email;

-- name: GetUserByID :one
SELECT id,
       email,
       first_name,
       last_name,
       phone_number,
       profile_image_url,
       date_of_birth,
       is_active,
       created_at,
       updated_at,
       last_login
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id,
       email,
       first_name,
       last_name,
       hashed_password,
       phone_number,
       profile_image_url,
       date_of_birth,
       is_active,
       created_at,
       updated_at,
       last_login
FROM users
WHERE email = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET first_name        = $2,
    last_name         = $3,
    phone_number      = $4,
    profile_image_url = $5,
    date_of_birth     = $6,
    updated_at        = NOW()
WHERE id = $1 RETURNING id, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login;

-- name: UpdateUserPassword :one
UPDATE users
SET hashed_password = $2,
    updated_at      = NOW()
WHERE id = $1 RETURNING id, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login;

-- name: UpdateUserLastLogin :one
UPDATE users
SET last_login = NOW()
WHERE id = $1 RETURNING id, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login;

-- name: DeactivateUser :one
UPDATE users
SET is_active  = FALSE,
    updated_at = NOW()
WHERE id = $1 RETURNING id, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login;

-- name: DeleteUser :exec
DELETE
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id,
       email,
       first_name,
       last_name,
       phone_number,
       profile_image_url,
       date_of_birth,
       is_active,
       created_at,
       updated_at,
       last_login
FROM users
ORDER BY created_at DESC LIMIT $1
OFFSET $2;

-- name: CountAllUsers :one
SELECT count(*) FROM users;

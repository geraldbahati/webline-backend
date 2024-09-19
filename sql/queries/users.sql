-- name: CreateUser :one
INSERT INTO users (
    id, email, hashed_password, first_name, last_name, phone_number,
    profile_image_url, date_of_birth, is_active, provider, provider_id
) VALUES (
             gen_random_uuid(), $1, $2, $3, $4, $5,
             $6, $7, true, $8, $9
         )
RETURNING id, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login, provider, provider_id;


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
       last_login,
         provider,
            provider_id,
            email_verified_at

FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, email, hashed_password, first_name, last_name, phone_number,
       profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login, provider, provider_id, email_verified_at
FROM users
WHERE email = $1;

-- name: GetUserByProvider :one
SELECT id, email, hashed_password, first_name, last_name, phone_number,
       profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login, provider, provider_id
FROM users
WHERE provider = $1 AND provider_id = $2;

-- name: UpdateUserEmailVerified :exec
UPDATE users
SET email_verified_at = NOW()
WHERE email = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET first_name        = $2,
    last_name         = $3,
    phone_number      = $4,
    profile_image_url = $5,
    date_of_birth     = $6,
    provider          = $7,
    provider_id       = $8,
    updated_at        = NOW()
WHERE id = $1 RETURNING id, email, first_name, last_name, phone_number, profile_image_url, date_of_birth, is_active, created_at, updated_at, last_login, provider, provider_id;

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

-- name: IsAdmin :one
SELECT EXISTS (
    SELECT 1
    FROM user_roles ur
             JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = $1 AND r.role_name = 'admin'
) AS is_admin;

-- name: MakeAdmin :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, (SELECT id FROM roles WHERE role_name = 'admin'))
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: GetUserProfileByID :one
SELECT
  u.id,
  u.email,
  u.profile_image_url,
  u.first_name,
  u.last_name,
  u.phone_number,
  u.date_of_birth,
  CASE WHEN ar.id IS NOT NULL THEN true ELSE false END AS request_admin,
  ar.reason AS admin_request_reason
FROM users u
LEFT JOIN admin_requests ar ON u.id = ar.user_id AND ar.status = 'PENDING'
WHERE u.id = $1;

-- name: UpdateUserInfo :exec
UPDATE users
SET
  profile_image_url = COALESCE($2, profile_image_url),
  first_name        = COALESCE($3, first_name),
  last_name         = COALESCE($4, last_name),
  phone_number      = COALESCE($5, phone_number),
  date_of_birth     = COALESCE($6, date_of_birth),
  updated_at        = now()
WHERE id = $1;

-- name: UpdateUser :exec
UPDATE users
SET
  first_name        = COALESCE($2, first_name),
  last_name         = COALESCE($3, last_name),
  phone_number      = COALESCE($4, phone_number),
  updated_at        = now()
WHERE id = $1;
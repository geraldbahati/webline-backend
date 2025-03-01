-- name: CreateCompany :one
INSERT INTO companies (
    name,
    kra_pin,
    address,
    phone_number,
    email
) VALUES (
    $1, -- name
    $2, -- kra_pin
    $3, -- address
    $4, -- phone_number
    $5 -- email
)
ON CONFLICT (kra_pin) DO UPDATE SET
    kra_pin = companies.kra_pin -- no-op update to return the existing row id
RETURNING id;

-- name: UpdateUserCompany :exec
UPDATE public.users
SET company_id = $2,
    updated_at = now()
WHERE id = $1;

-- name: GetCompany :one
SELECT id
FROM companies
WHERE name = $1
AND kra_pin = $2;

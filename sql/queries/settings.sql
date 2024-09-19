-- name: GetVATPercentage :one
SELECT vat_percentage FROM settings WHERE id = TRUE;

-- name: UpdateVATPercentage :exec
UPDATE settings SET vat_percentage = $1 WHERE id = TRUE;

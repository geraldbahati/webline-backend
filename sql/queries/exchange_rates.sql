-- name: GetLatestExchangeRate :one
SELECT rate_to_kes
FROM exchange_rates
WHERE currency_code = $1
  AND valid_from <= NOW()
  AND (valid_to IS NULL OR valid_to >= NOW())
ORDER BY valid_from DESC
LIMIT 1;

-- name: InsertExchangeRate :exec
INSERT INTO exchange_rates (currency_code, rate_to_kes, valid_from, valid_to)
VALUES ($1, $2, $3, $4);

-- name: UpdateExchangeRate :exec
UPDATE exchange_rates
SET rate_to_kes = $2, valid_from = $3, valid_to = $4
WHERE currency_code = $1
  AND valid_from <= NOW()
  AND (valid_to IS NULL OR valid_to >= NOW());

-- name: DeleteExchangeRate :exec
DELETE FROM exchange_rates
WHERE currency_code = $1
  AND valid_from = $2
  AND valid_to = $3;

-- name: GetAllExchangeRatesForCurrency :many
SELECT id, currency_code, rate_to_kes, valid_from, valid_to
FROM exchange_rates
WHERE currency_code = $1
ORDER BY valid_from DESC;

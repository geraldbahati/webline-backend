-- name: GetUserRecommendations :many
SELECT
    r.product_id,
    p.name,
    p.description,
    p.usd_price,
    r.score
FROM
    recommendations r
        JOIN
    products p ON r.product_id = p.id
WHERE
    r.user_id = $1
ORDER BY
    r.score DESC
    LIMIT $2;

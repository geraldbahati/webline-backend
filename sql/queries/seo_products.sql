-- name: UpdateProductSEO :exec
UPDATE products
SET
    part_number = $2,
    meta_title = $3,
    meta_description = $4,
    meta_keywords = $5,
    updated_at = now()
WHERE id = $1;

-- name: GetProductSEO :one
SELECT
    p.id,
    p.part_number,
    p.meta_title,
    p.meta_description,
    p.meta_keywords,
    p.price,
    c.name AS brand_name,
    (SELECT image_url FROM product_images WHERE product_id = p.id ORDER BY created_at LIMIT 1) AS image_url
FROM products p
         JOIN categories c ON p.category_id = c.id
WHERE p.id = $1;

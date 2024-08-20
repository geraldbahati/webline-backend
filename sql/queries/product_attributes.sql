-- name: CreateProductAttribute :one
INSERT INTO product_attributes (name, attribute_type_id)
VALUES ($1, (SELECT at.id FROM attribute_types at WHERE at.name = $2))
RETURNING id;


-- name: CreateProductAttributeValue :one
INSERT INTO product_attribute_values (attribute_id, value, hex_value, category_id)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: CreateProductToAttributeValue :exec
INSERT INTO product_to_attribute_values (
    product_id, attribute_value_id
) VALUES (
             $1, $2
         );

-- name: GetProductAttributesWithValues :many
SELECT
    pa.id AS attribute_id,
    pa.name AS attribute_name,
    COALESCE(pav.value, '') AS attribute_value,
    COALESCE(pav.hex_value, '') AS hex_value
FROM
    product_attributes pa
        LEFT JOIN
    product_attribute_values pav ON pa.id = pav.attribute_id
WHERE
    (pav.category_id = $1 OR pav.category_id IS NULL)
ORDER BY
    pa.name, pav.value;


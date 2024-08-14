-- name: CreateProductAttribute :one
INSERT INTO product_attributes (name, attribute_type)
VALUES ($1, $2)
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

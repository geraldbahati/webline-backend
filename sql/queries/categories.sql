-- name: CreateCategory :one
INSERT INTO categories (name, parent_id, position)
VALUES ($1, $2, $3)
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position;


-- name: GetCategoryByID :one
SELECT id, name, parent_id, created_at, updated_at, is_active, position
FROM categories
WHERE id = $1;

-- name: ListCategories :many
SELECT id, name, parent_id, created_at, updated_at, is_active, position
FROM categories
ORDER BY position;

-- name: UpdateCategory :one
UPDATE categories
SET name = $2, parent_id = $3, position = $4, updated_at = NOW()
WHERE id = $1
    RETURNING id, name, parent_id, created_at, updated_at, is_active, position;

-- name: SoftDeleteCategory :exec
DELETE FROM categories
WHERE id = $1;

-- name: HardDeleteCategory :exec
DELETE FROM categories
WHERE id = $1;

-- name: GetCategoriesByParentID :many
SELECT id, name, parent_id, created_at, updated_at, is_active, position
FROM categories
WHERE parent_id = $1
ORDER BY position;

-- name: GetCategoriesWithProductsCount :many
SELECT c.id, c.name, c.parent_id, c.position, c.created_at, c.updated_at, c.is_active,COUNT(p.id) as products_count
FROM categories c
         LEFT JOIN products p ON c.id = p.category_id
GROUP BY c.id
ORDER BY c.position;

-- name: GetCategoryTree :many
WITH RECURSIVE category_tree AS (
    SELECT id, name, parent_id, created_at, updated_at, is_active, position
    FROM categories
    WHERE parent_id IS NULL
    UNION
    SELECT c.id, c.name, c.parent_id, c.created_at, c.updated_at, c.is_active, c.position
    FROM categories c
             INNER JOIN category_tree ct ON ct.id = c.parent_id
)
SELECT id, name, parent_id, created_at, updated_at, is_active, position
FROM category_tree
ORDER BY position, parent_id, name;

-- name: CheckCategoryExistence :one
SELECT EXISTS (
    SELECT 1
    FROM categories
    WHERE id = $1
) AS exists;

-- name: GetCategoriesWithSubcategoryCount :many
SELECT c.id, c.name, c.parent_id, c.created_at, c.position ,c.updated_at, c.is_active,COUNT(sc.id) as subcategory_count
FROM categories c
         LEFT JOIN categories sc ON c.id = sc.parent_id
GROUP BY c.id
ORDER BY c.position;

-- name: GetParentCategories :many
SELECT id, name, parent_id, created_at, updated_at, is_active, position
FROM categories
WHERE parent_id IS NULL
ORDER BY position;

-- name: GetCategoryByName :one
SELECT id, name, parent_id, created_at, updated_at, is_active, position
FROM categories
WHERE name = $1;

-- name: GetCategoryHierarchy :many
WITH RECURSIVE category_hierarchy AS (
    SELECT
        c1.id AS category_id,
        c1.name AS category_name,
        c1.parent_id,
        c1.position
    FROM categories c1
    WHERE c1.parent_id IS NULL
    UNION ALL
    SELECT
        c2.id AS category_id,
        c2.name AS category_name,
        c2.parent_id,
        c2.position
    FROM categories c2
    INNER JOIN category_hierarchy ch ON ch.category_id = c2.parent_id
),
product_details AS (
    SELECT
        p.id AS product_id,
        p.category_id,
        p.name AS product_name,
        proc.name AS processor,
        sz.size AS size,
        st.name AS storage,
        (SELECT pi.image_url FROM product_images pi WHERE pi.product_id = p.id ORDER BY pi.created_at ASC LIMIT 1) AS image_url,
        COALESCE(SUM(oi.quantity), 0) AS total_sold,
        ROW_NUMBER() OVER (PARTITION BY p.category_id ORDER BY COALESCE(SUM(oi.quantity), 0) DESC) AS rank
    FROM
        products p
    LEFT JOIN
        product_processors pp ON pp.product_id = p.id
    LEFT JOIN
        processors proc ON proc.id = pp.processor_id
    LEFT JOIN
        product_sizes ps ON p.id = ps.product_id
    LEFT JOIN
        sizes sz ON ps.size_id = sz.id
    LEFT JOIN
        product_storage_options pso ON p.id = pso.product_id
    LEFT JOIN
        storage_options st ON st.id = pso.storage_option_id
    LEFT JOIN
        order_items oi ON p.id = oi.product_id
    WHERE
        p.is_active = true
    GROUP BY
        p.id, p.category_id, p.name, proc.name, sz.size, st.name
),
category_details AS (
    SELECT
        ch.category_id,
        ch.category_name,
        ch.parent_id,
        ch.position,
        pd.product_id,
        pd.product_name,
        pd.processor,
        pd.size,
        pd.storage,
        pd.image_url,
        pd.total_sold
    FROM
        category_hierarchy ch
    LEFT JOIN
        product_details pd ON pd.category_id = ch.category_id
    WHERE
        pd.rank <= 3 OR pd.rank IS NULL
)
SELECT
    ch.category_id,
    ch.category_name,
    ch.parent_id,
    ch.position,
    cd.product_id,
    cd.product_name,
    cd.processor,
    cd.size,
    cd.storage,
    cd.image_url,
    cd.total_sold
FROM
    category_hierarchy ch
LEFT JOIN
    category_details cd ON cd.category_id = ch.category_id
ORDER BY
    ch.position;

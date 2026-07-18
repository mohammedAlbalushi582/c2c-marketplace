-- name: ListCategories :many
SELECT * FROM categories
WHERE is_active = true
ORDER BY display_order, name_ar;

-- name: ListAllCategories :many
SELECT * FROM categories
ORDER BY display_order, name_ar;

-- name: GetCategoryByID :one
SELECT * FROM categories WHERE id = $1;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: CreateCategory :one
INSERT INTO categories (parent_id, slug, name_ar, name_en, icon, display_order, is_active)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: UpdateCategory :one
UPDATE categories
SET parent_id = $2, slug = $3, name_ar = $4, name_en = $5, icon = $6,
    display_order = $7, is_active = $8
WHERE id = $1
RETURNING *;

-- name: DeleteCategory :exec
DELETE FROM categories WHERE id = $1;

-- name: CreateCategoryField :one
INSERT INTO category_fields
    (category_id, field_key, label_ar, label_en, field_type, unit, options,
     is_required, is_filterable, display_order)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: ListFieldsByCategory :many
SELECT * FROM category_fields
WHERE category_id = $1
ORDER BY display_order;

-- Resolve fields for a category AND all its ancestors (inherited attributes).
-- name: ListResolvedFieldsForCategory :many
WITH RECURSIVE chain(id, parent_id) AS (
    SELECT c.id, c.parent_id FROM categories c WHERE c.id = $1
    UNION ALL
    SELECT c.id, c.parent_id FROM categories c
    JOIN chain ch ON c.id = ch.parent_id
)
SELECT f.* FROM category_fields f
JOIN chain ch ON f.category_id = ch.id
ORDER BY f.display_order;

-- name: DeleteCategoryField :exec
DELETE FROM category_fields WHERE id = $1;

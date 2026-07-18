-- name: CreateListing :one
INSERT INTO listings
    (user_id, category_id, location_id, title, slug, description, price, currency,
     price_type, contact_phone, whatsapp_number, status, expires_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetListingByID :one
SELECT sqlc.embed(l),
       c.slug AS category_slug, c.name_ar AS category_name_ar, c.name_en AS category_name_en,
       loc.name_ar AS location_name_ar, loc.name_en AS location_name_en
FROM listings l
JOIN categories c ON c.id = l.category_id
LEFT JOIN locations loc ON loc.id = l.location_id
WHERE l.id = $1 AND l.deleted_at IS NULL;

-- name: UpdateListing :one
UPDATE listings
SET category_id = $2, location_id = $3, title = $4, slug = $5, description = $6,
    price = $7, price_type = $8, contact_phone = $9, whatsapp_number = $10
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateListingStatus :one
UPDATE listings
SET status = $2,
    published_at = CASE WHEN $2 = 'active'::listing_status AND published_at IS NULL THEN now() ELSE published_at END
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteListing :exec
UPDATE listings SET deleted_at = now() WHERE id = $1;

-- name: IncrementListingViews :exec
UPDATE listings SET views_count = views_count + 1 WHERE id = $1;

-- name: ListListingsByUser :many
SELECT l.*, COALESCE(pi.storage_key, '') AS primary_image
FROM listings l
LEFT JOIN LATERAL (
    SELECT storage_key FROM listing_images i
    WHERE i.listing_id = l.id ORDER BY i.is_primary DESC, i.display_order LIMIT 1
) pi ON true
WHERE l.user_id = $1 AND l.deleted_at IS NULL
ORDER BY l.created_at DESC;

-- name: SearchListings :many
SELECT l.*,
       c.slug AS category_slug, c.name_ar AS category_name_ar,
       loc.name_ar AS location_name_ar,
       COALESCE(pi.storage_key, '') AS primary_image
FROM listings l
JOIN categories c ON c.id = l.category_id
LEFT JOIN locations loc ON loc.id = l.location_id
LEFT JOIN LATERAL (
    SELECT storage_key FROM listing_images i
    WHERE i.listing_id = l.id ORDER BY i.is_primary DESC, i.display_order LIMIT 1
) pi ON true
WHERE l.deleted_at IS NULL
  AND l.status = 'active'
  AND (sqlc.narg('category_id')::bigint IS NULL OR l.category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('location_id')::bigint IS NULL OR l.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('min_price')::numeric IS NULL OR l.price >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::numeric IS NULL OR l.price <= sqlc.narg('max_price'))
  AND (sqlc.narg('keyword')::text IS NULL
       OR l.title ILIKE '%' || sqlc.narg('keyword') || '%'
       OR l.description ILIKE '%' || sqlc.narg('keyword') || '%')
ORDER BY l.is_featured DESC, l.published_at DESC NULLS LAST, l.created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountListings :one
SELECT count(*) FROM listings l
WHERE l.deleted_at IS NULL
  AND l.status = 'active'
  AND (sqlc.narg('category_id')::bigint IS NULL OR l.category_id = sqlc.narg('category_id'))
  AND (sqlc.narg('location_id')::bigint IS NULL OR l.location_id = sqlc.narg('location_id'))
  AND (sqlc.narg('min_price')::numeric IS NULL OR l.price >= sqlc.narg('min_price'))
  AND (sqlc.narg('max_price')::numeric IS NULL OR l.price <= sqlc.narg('max_price'))
  AND (sqlc.narg('keyword')::text IS NULL
       OR l.title ILIKE '%' || sqlc.narg('keyword') || '%'
       OR l.description ILIKE '%' || sqlc.narg('keyword') || '%');

-- ---- images ----
-- name: AddListingImage :one
INSERT INTO listing_images (listing_id, storage_key, alt_text, is_primary, display_order)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: ListImagesByListing :many
SELECT * FROM listing_images
WHERE listing_id = $1
ORDER BY is_primary DESC, display_order;

-- name: GetListingImage :one
SELECT * FROM listing_images WHERE id = $1;

-- name: DeleteListingImage :exec
DELETE FROM listing_images WHERE id = $1;

-- ---- attributes (EAV) ----
-- name: UpsertListingAttribute :exec
INSERT INTO listing_attributes
    (listing_id, field_id, value_text, value_number, value_boolean, value_date, value_json)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (listing_id, field_id) DO UPDATE
SET value_text = EXCLUDED.value_text,
    value_number = EXCLUDED.value_number,
    value_boolean = EXCLUDED.value_boolean,
    value_date = EXCLUDED.value_date,
    value_json = EXCLUDED.value_json;

-- name: ListAttributesByListing :many
SELECT a.*, f.field_key, f.label_ar, f.label_en, f.field_type, f.unit, f.options
FROM listing_attributes a
JOIN category_fields f ON f.id = a.field_id
WHERE a.listing_id = $1
ORDER BY f.display_order;

-- name: DeleteAttributesByListing :exec
DELETE FROM listing_attributes WHERE listing_id = $1;

-- name: AdminListListings :many
SELECT l.*, COALESCE(pi.storage_key, '') AS primary_image,
       c.name_ar AS category_name_ar
FROM listings l
JOIN categories c ON c.id = l.category_id
LEFT JOIN LATERAL (
    SELECT storage_key FROM listing_images i
    WHERE i.listing_id = l.id ORDER BY i.is_primary DESC, i.display_order LIMIT 1
) pi ON true
WHERE l.deleted_at IS NULL
  AND (sqlc.narg('status')::listing_status IS NULL OR l.status = sqlc.narg('status'))
ORDER BY l.created_at DESC
LIMIT 100;

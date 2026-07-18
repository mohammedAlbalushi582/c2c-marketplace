-- name: ListLocations :many
SELECT * FROM locations
ORDER BY type, name_ar;

-- name: ListGovernorates :many
SELECT * FROM locations
WHERE type = 'governorate'
ORDER BY name_ar;

-- name: ListWilayatsByGovernorate :many
SELECT * FROM locations
WHERE parent_id = $1
ORDER BY name_ar;

-- name: GetLocationByID :one
SELECT * FROM locations WHERE id = $1;

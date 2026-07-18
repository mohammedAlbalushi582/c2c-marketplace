-- name: AddFavorite :exec
INSERT INTO favorites (user_id, listing_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveFavorite :exec
DELETE FROM favorites WHERE user_id = $1 AND listing_id = $2;

-- name: IsFavorite :one
SELECT EXISTS (
    SELECT 1 FROM favorites WHERE user_id = $1 AND listing_id = $2
);

-- name: ListFavoritesByUser :many
SELECT l.*, COALESCE(pi.storage_key, '') AS primary_image
FROM favorites fv
JOIN listings l ON l.id = fv.listing_id
LEFT JOIN LATERAL (
    SELECT storage_key FROM listing_images i
    WHERE i.listing_id = l.id ORDER BY i.is_primary DESC, i.display_order LIMIT 1
) pi ON true
WHERE fv.user_id = $1 AND l.deleted_at IS NULL
ORDER BY fv.created_at DESC;

-- name: CreateReport :one
INSERT INTO reports (listing_id, reporter_user_id, reason, details)
VALUES ($1, $2, $3, $4)
RETURNING *;

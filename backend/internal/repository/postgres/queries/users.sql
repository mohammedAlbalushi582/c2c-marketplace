-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, full_name, phone, whatsapp_number)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT * FROM users WHERE username = $1;

-- name: UpdateUserProfile :one
UPDATE users
SET full_name = $2, phone = $3, whatsapp_number = $4
WHERE id = $1
RETURNING *;

-- ---- admin ----

-- name: AdminListUsers :many
SELECT u.*,
       (SELECT count(*) FROM listings l
        WHERE l.user_id = u.id AND l.deleted_at IS NULL) AS listings_count
FROM users u
WHERE u.status <> 'deleted'
  AND (sqlc.narg('role')::user_role IS NULL OR u.role = sqlc.narg('role'))
  AND (sqlc.narg('keyword')::text IS NULL
       OR u.email ILIKE '%' || sqlc.narg('keyword') || '%'
       OR u.full_name ILIKE '%' || sqlc.narg('keyword') || '%'
       OR u.username ILIKE '%' || sqlc.narg('keyword') || '%')
ORDER BY u.created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountUsers :one
SELECT count(*) FROM users u
WHERE u.status <> 'deleted'
  AND (sqlc.narg('role')::user_role IS NULL OR u.role = sqlc.narg('role'))
  AND (sqlc.narg('keyword')::text IS NULL
       OR u.email ILIKE '%' || sqlc.narg('keyword') || '%'
       OR u.full_name ILIKE '%' || sqlc.narg('keyword') || '%'
       OR u.username ILIKE '%' || sqlc.narg('keyword') || '%');

-- name: AdminSetUserRole :one
UPDATE users SET role = $2 WHERE id = $1 RETURNING *;

-- name: AdminSetUserStatus :one
UPDATE users SET status = $2 WHERE id = $1 RETURNING *;

-- name: CountAdmins :one
SELECT count(*) FROM users WHERE role = 'admin' AND status = 'active';

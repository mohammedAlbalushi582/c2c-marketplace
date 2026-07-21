-- name: CreateContactMessage :one
INSERT INTO contact_messages (name, phone, email, message)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListContactMessages :many
SELECT * FROM contact_messages
ORDER BY is_read, created_at DESC
LIMIT sqlc.arg('lim') OFFSET sqlc.arg('off');

-- name: CountContactMessages :one
SELECT count(*) FROM contact_messages;

-- name: CountUnreadContactMessages :one
SELECT count(*) FROM contact_messages WHERE is_read = false;

-- name: MarkContactMessageRead :exec
UPDATE contact_messages SET is_read = true WHERE id = $1;

-- name: DeleteContactMessage :exec
DELETE FROM contact_messages WHERE id = $1;

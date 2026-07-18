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

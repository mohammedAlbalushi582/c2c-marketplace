-- name: GetSetting :one
SELECT value FROM app_settings WHERE key = $1;

-- name: ListSettings :many
SELECT * FROM app_settings ORDER BY key;

-- name: UpsertSetting :exec
INSERT INTO app_settings (key, value, updated_at)
VALUES ($1, $2, now())
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = now();

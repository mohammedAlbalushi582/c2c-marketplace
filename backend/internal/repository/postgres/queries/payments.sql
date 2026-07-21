-- name: CreatePayment :one
INSERT INTO payments (user_id, listing_id, purpose, amount, currency, provider)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetPayment :one
SELECT * FROM payments WHERE id = $1;

-- name: GetPaymentByProviderRef :one
SELECT * FROM payments WHERE provider_ref = $1;

-- name: SetPaymentCheckout :one
-- Records the gateway session/checkout id after a checkout is created.
UPDATE payments
SET provider_ref = $2, provider_payload = $3
WHERE id = $1
RETURNING *;

-- name: MarkPaymentPaid :one
-- Idempotent: only a still-pending payment transitions to paid.
UPDATE payments
SET status = 'paid', paid_at = now(), provider_payload = COALESCE($2, provider_payload)
WHERE id = $1 AND status = 'pending'
RETURNING *;

-- name: MarkPaymentFailed :exec
UPDATE payments
SET status = 'failed', provider_payload = COALESCE($2, provider_payload)
WHERE id = $1 AND status = 'pending';

-- name: ListPaymentsByUser :many
SELECT * FROM payments
WHERE user_id = $1
ORDER BY created_at DESC;

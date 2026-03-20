-- name: GetClientByID :one
SELECT id, clerk_user_id,
       first_name, last_name, email, phone, notes, can_pay_later,
       created_at, updated_at
FROM client.clients
WHERE id = $1;

-- name: GetClientByClerkUserID :one
SELECT id, clerk_user_id,
       first_name, last_name, email, phone, notes, can_pay_later,
       created_at, updated_at
FROM client.clients
WHERE clerk_user_id = $1;

-- name: CreateClient :exec
INSERT INTO client.clients (
    id, clerk_user_id,
    first_name, last_name, email, phone, notes, can_pay_later,
    created_at, updated_at
) VALUES (
    $1, $2,
    $3, $4, $5, $6, $7, $8,
    $9, $10
);

-- name: UpdateClient :execrows
UPDATE client.clients
SET
    phone         = COALESCE(sqlc.narg('phone'), phone),
    notes         = COALESCE(sqlc.narg('notes'), notes),
    can_pay_later = COALESCE(sqlc.narg('can_pay_later'), can_pay_later),
    updated_at    = $2
WHERE id = $1;
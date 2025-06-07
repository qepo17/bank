-- name: CreateAccount :one
INSERT INTO accounts (id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id, created_at, updated_at;

-- name: GetAccountByID :one
SELECT id, created_at, updated_at
FROM accounts
WHERE id = $1;
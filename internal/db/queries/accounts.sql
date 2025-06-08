-- name: CreateAccount :one
INSERT INTO accounts (id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id, created_at, updated_at;

-- name: GetAccountByID :one
SELECT id, created_at, updated_at
FROM accounts
WHERE id = $1;

-- name: GetAccountBalanceByAccountID :one
SELECT get_account_balance($1, $2);

-- name: CheckAccountExists :one
SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1);
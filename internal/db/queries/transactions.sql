-- name: CreateCreditTransaction :one
INSERT INTO transactions (account_id, transfer_id, amount, trx_type, created_at)
VALUES ($1, $2, $3, 'CREDIT', NOW())
RETURNING id, account_id, transfer_id, amount, trx_type, created_at;

-- name: CreateDebitTransaction :one
INSERT INTO transactions (account_id, transfer_id, amount, trx_type, created_at)
VALUES ($1, $2, $3, 'DEBIT', NOW())
RETURNING id, account_id, transfer_id, amount, trx_type, created_at;

-- name: CreateTransferTransaction :one
SELECT transfer_funds($1, $2, $3) as result;
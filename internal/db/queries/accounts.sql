-- name: CreateAccount :one
INSERT INTO accounts (id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id, created_at, updated_at;

-- name: GetAccountByID :one
SELECT id, created_at, updated_at
FROM accounts
WHERE id = $1;

-- name: GetAccountBalanceByAccountID :one
WITH filters AS (
    SELECT
        $1 as account_id
),
latest_snapshot AS (
    SELECT 
        account_balance_snapshots.account_id,
        balance,
        last_transaction_id,
        created_at
    FROM account_balance_snapshots
    JOIN filters ON account_balance_snapshots.account_id = filters.account_id
    ORDER BY created_at DESC
    LIMIT 1
),
recent_transactions AS (
    SELECT 
        COALESCE(
            SUM(CASE 
                WHEN t.trx_type = 'CREDIT' THEN t.amount
                WHEN t.trx_type = 'DEBIT' THEN -t.amount
                ELSE 0
            END), 0
        ) as transaction_delta
    FROM transactions t
    CROSS JOIN latest_snapshot ls
    JOIN filters ON t.account_id = filters.account_id
      AND t.id > ls.last_transaction_id
)
SELECT
    ls.balance + COALESCE(rt.transaction_delta, 0) as current_balance
FROM latest_snapshot ls
CROSS JOIN recent_transactions rt;
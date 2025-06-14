// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: accounts.sql

package sqlc

import (
	"context"
)

const checkAccountExists = `-- name: CheckAccountExists :one
SELECT EXISTS(SELECT 1 FROM accounts WHERE id = $1)
`

func (q *Queries) CheckAccountExists(ctx context.Context, id int64) (bool, error) {
	row := q.db.QueryRowContext(ctx, checkAccountExists, id)
	var exists bool
	err := row.Scan(&exists)
	return exists, err
}

const createAccount = `-- name: CreateAccount :one
INSERT INTO accounts (id, created_at, updated_at)
VALUES ($1, NOW(), NOW())
RETURNING id, created_at, updated_at
`

func (q *Queries) CreateAccount(ctx context.Context, id int64) (Account, error) {
	row := q.db.QueryRowContext(ctx, createAccount, id)
	var i Account
	err := row.Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

const getAccountBalanceByAccountID = `-- name: GetAccountBalanceByAccountID :one
SELECT get_account_balance($1, $2)
`

type GetAccountBalanceByAccountIDParams struct {
	FilterAccountID     int64 `db:"filter_account_id" json:"filter_account_id"`
	FilterLockForUpdate bool  `db:"filter_lock_for_update" json:"filter_lock_for_update"`
}

func (q *Queries) GetAccountBalanceByAccountID(ctx context.Context, arg GetAccountBalanceByAccountIDParams) (string, error) {
	row := q.db.QueryRowContext(ctx, getAccountBalanceByAccountID, arg.FilterAccountID, arg.FilterLockForUpdate)
	var get_account_balance string
	err := row.Scan(&get_account_balance)
	return get_account_balance, err
}

const getAccountByID = `-- name: GetAccountByID :one
SELECT id, created_at, updated_at
FROM accounts
WHERE id = $1
`

func (q *Queries) GetAccountByID(ctx context.Context, id int64) (Account, error) {
	row := q.db.QueryRowContext(ctx, getAccountByID, id)
	var i Account
	err := row.Scan(&i.ID, &i.CreatedAt, &i.UpdatedAt)
	return i, err
}

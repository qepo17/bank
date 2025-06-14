// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package sqlc

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

type Account struct {
	ID        int64        `db:"id" json:"id"`
	CreatedAt sql.NullTime `db:"created_at" json:"created_at"`
	UpdatedAt sql.NullTime `db:"updated_at" json:"updated_at"`
}

type AccountBalanceSnapshot struct {
	ID                int64        `db:"id" json:"id"`
	AccountID         int64        `db:"account_id" json:"account_id"`
	Balance           string       `db:"balance" json:"balance"`
	LastTransactionID int64        `db:"last_transaction_id" json:"last_transaction_id"`
	CreatedAt         sql.NullTime `db:"created_at" json:"created_at"`
}

type Transaction struct {
	ID         int64           `db:"id" json:"id"`
	AccountID  int64           `db:"account_id" json:"account_id"`
	TransferID sql.NullInt64   `db:"transfer_id" json:"transfer_id"`
	Amount     decimal.Decimal `db:"amount" json:"amount"`
	TrxType    string          `db:"trx_type" json:"trx_type"`
	CreatedAt  sql.NullTime    `db:"created_at" json:"created_at"`
}

type Transfer struct {
	ID            int64        `db:"id" json:"id"`
	FromAccountID int64        `db:"from_account_id" json:"from_account_id"`
	ToAccountID   int64        `db:"to_account_id" json:"to_account_id"`
	CreatedAt     sql.NullTime `db:"created_at" json:"created_at"`
}

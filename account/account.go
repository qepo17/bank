package account

import (
	"bank/entity"
	"bank/internal/db/sqlc"
	"context"
	"database/sql"
	"errors"
)

type AccountDomain struct {
	db      *sql.DB
	queries *sqlc.Queries
}

func NewAccountDomain(db *sql.DB, sqlc *sqlc.Queries) (*AccountDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if sqlc == nil {
		return nil, errors.New("sqlc is nil")
	}

	return &AccountDomain{
		db:      db,
		queries: sqlc,
	}, nil
}

// CreateAccount creates a new account and initial balance transaction atomically
func (d *AccountDomain) CreateAccount(ctx context.Context, account *entity.CreateAccount) error {
	if err := account.Validate(); err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := d.queries.WithTx(tx)

	_, err = qtx.CreateAccount(ctx, int64(account.AccountID))
	if err != nil {
		return err
	}

	_, err = qtx.CreateCreditTransaction(ctx, sqlc.CreateCreditTransactionParams{
		AccountID:  int64(account.AccountID),
		TransferID: sql.NullInt64{Valid: false},
		Amount:     account.InitialBalance,
	})
	if err != nil {
		return err
	}

	return tx.Commit()
}

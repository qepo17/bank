package account

import (
	"bank/entity"
	"bank/internal/db/sqlc"
	"bank/internal/logger"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

type AccountDomain struct {
	db      *sql.DB
	queries *sqlc.Queries
	logger  *logger.Logger
}

func NewAccountDomain(db *sql.DB, sqlc *sqlc.Queries, logger *logger.Logger) (*AccountDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if sqlc == nil {
		return nil, errors.New("sqlc is nil")
	}

	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	log := logger.WithField("domain", "account")

	return &AccountDomain{
		db:      db,
		queries: sqlc,
		logger:  log,
	}, nil
}

// CreateAccount creates a new account and initial balance transaction atomically
func (d *AccountDomain) CreateAccount(ctx context.Context, account *entity.CreateAccount) error {
	if err := account.Validate(); err != nil {
		return err
	}

	tx, err := d.db.Begin()
	if err != nil {
		d.logger.Error(ctx, "failed to begin transaction for account_id=%d: %v", account.AccountID, err)
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()
	qtx := d.queries.WithTx(tx)

	_, err = qtx.CreateAccount(ctx, int64(account.AccountID))
	if err != nil {
		d.logger.Error(ctx, "failed to create account record for account_id=%d: %v", account.AccountID, err)
		return fmt.Errorf("failed to create account: %w", err)
	}

	noTransferID := sql.NullInt64{Valid: false}
	_, err = qtx.CreateCreditTransaction(ctx, sqlc.CreateCreditTransactionParams{
		AccountID:  int64(account.AccountID),
		TransferID: noTransferID,
		Amount:     account.InitialBalance,
	})
	if err != nil {
		d.logger.Error(ctx, "failed to create initial credit transaction for account_id=%d: %v", account.AccountID, err)
		return fmt.Errorf("failed to create initial transaction: %w", err)
	}

	if err := tx.Commit(); err != nil {
		d.logger.Error(ctx, "failed to commit transaction for account_id=%d: %v", account.AccountID, err)
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (d *AccountDomain) GetAccountBalance(ctx context.Context, accountID uint64) (decimal.Decimal, error) {
	exists, err := d.queries.CheckAccountExists(ctx, int64(accountID))
	if err != nil {
		return decimal.Zero, fmt.Errorf("failed to check account exists: %w", err)
	}

	if !exists {
		return decimal.Zero, entity.ErrNoRows
	}

	balance, err := d.queries.GetAccountBalanceByAccountID(ctx, sqlc.GetAccountBalanceByAccountIDParams{
		FilterAccountID:     int64(accountID),
		FilterLockForUpdate: true,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return decimal.Zero, entity.ErrNoRows
		}

		d.logger.Error(ctx, "failed to get account balance for account_id=%d: %v", accountID, err)
		return decimal.Zero, fmt.Errorf("failed to get account balance: %w", err)
	}
	return decimal.NewFromString(balance)
}

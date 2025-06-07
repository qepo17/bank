package account

import (
	"bank/internal/repository"
	"database/sql"
	"errors"
)

type AccountDomain struct {
	db                *sql.DB
	accountRepository *repository.AccountRepository
}

func NewAccountDomain(db *sql.DB, accountRepository *repository.AccountRepository) (*AccountDomain, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	if accountRepository == nil {
		return nil, errors.New("account repository is nil")
	}

	return &AccountDomain{
		db:                db,
		accountRepository: accountRepository,
	}, nil
}

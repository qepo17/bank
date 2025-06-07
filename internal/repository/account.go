package repository

import (
	"database/sql"
	"errors"
)

type AccountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) (*AccountRepository, error) {
	if db == nil {
		return nil, errors.New("db is nil")
	}

	return &AccountRepository{db: db}, nil
}

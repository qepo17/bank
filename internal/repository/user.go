package repository

import (
	"bank/entity"
	"context"
	"database/sql"
	"errors"
)

type CustomerRepository struct {
	db *sql.DB
}

func NewCustomerRepository(db *sql.DB) (*CustomerRepository, error) {
	if db == nil {
		return nil, errors.New("missing db")
	}

	return &CustomerRepository{
		db: db,
	}, nil
}

func (s *CustomerRepository) FindByID(ctx context.Context, id int) (entity.Customer, error) {
	return entity.Customer{}, nil
}

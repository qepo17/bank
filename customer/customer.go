package customer

import (
	"bank/entity"
	"bank/internal/repository"
	"context"
	"database/sql"
	"errors"
)

type CustomerDomain struct {
	db *sql.DB

	customerRepository *repository.CustomerRepository
}

func NewCustomerDomain(db *sql.DB, customerRepository *repository.CustomerRepository) (*CustomerDomain, error) {
	if db == nil {
		return nil, errors.New("missing db")
	}

	if customerRepository == nil {
		return nil, errors.New("missing customer repository")
	}

	return &CustomerDomain{
		db:                 db,
		customerRepository: customerRepository,
	}, nil
}

func (s *CustomerDomain) FindByID(ctx context.Context, id int) (entity.Customer, error) {
	return s.customerRepository.FindByID(ctx, id)
}

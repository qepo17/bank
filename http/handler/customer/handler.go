package customer

import (
	customer "bank/customer"
	"errors"
)

type Handler struct {
	customerDomain *customer.CustomerDomain
}

func NewHandler(customerService *customer.CustomerDomain) (*Handler, error) {
	if customerService == nil {
		return nil, errors.New("customer domain is nil")
	}

	return &Handler{
		customerDomain: customerService,
	}, nil
}

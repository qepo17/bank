package customer

import (
	"bank/account"
	"errors"
)

type Handler struct {
	accountDomain *account.AccountDomain
}

func NewHandler(accountDomain *account.AccountDomain) (*Handler, error) {
	if accountDomain == nil {
		return nil, errors.New("account domain is nil")
	}

	return &Handler{
		accountDomain: accountDomain,
	}, nil
}

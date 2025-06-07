package customer

import (
	"bank/account"
	"bank/internal/logger"
	"errors"
)

type Handler struct {
	accountDomain *account.AccountDomain
	logger        *logger.Logger
}

func NewHandler(accountDomain *account.AccountDomain, logger *logger.Logger) (*Handler, error) {
	if accountDomain == nil {
		return nil, errors.New("account domain is nil")
	}

	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	log := logger.WithField("handler", "customer")
	return &Handler{
		accountDomain: accountDomain,
		logger:        log,
	}, nil
}

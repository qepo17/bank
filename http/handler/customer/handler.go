package customer

import (
	"bank/account"
	"bank/internal/logger"
	"bank/transaction"
	"errors"
)

type Handler struct {
	accountDomain     *account.AccountDomain
	logger            *logger.Logger
	transactionDomain *transaction.TransactionDomain
}

func NewHandler(accountDomain *account.AccountDomain, transactionDomain *transaction.TransactionDomain, logger *logger.Logger) (*Handler, error) {
	if accountDomain == nil {
		return nil, errors.New("account domain is nil")
	}

	if transactionDomain == nil {
		return nil, errors.New("transaction domain is nil")
	}

	if logger == nil {
		return nil, errors.New("logger is nil")
	}

	log := logger.WithField("handler", "customer")
	return &Handler{
		accountDomain:     accountDomain,
		logger:            log,
		transactionDomain: transactionDomain,
	}, nil
}

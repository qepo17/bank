package entity

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type AccountType string

const (
	AccountTypeSavings AccountType = "SAVINGS"
	AccountTypeCredit  AccountType = "CREDIT"
)

type CurrencyCode string

const (
	CurrencyCodeUSD CurrencyCode = "USD"
	CurrencyCodeEUR CurrencyCode = "EUR"
)

type Account struct {
	ModelWithUpdatedAt
}

type CreateAccount struct {
	AccountID      uint64
	InitialBalance decimal.Decimal
}

func (a *CreateAccount) Validate() error {
	msgs := []string{}
	if a.AccountID == 0 {
		msgs = append(msgs, "account id is required")
	}
	if a.InitialBalance.IsZero() {
		msgs = append(msgs, "initial balance is required")
	}
	if a.InitialBalance.LessThan(decimal.Zero) {
		msgs = append(msgs, "initial balance must be greater than 0")
	}
	if len(msgs) > 0 {
		return fmt.Errorf("%w: %s", ErrValidation, strings.Join(msgs, ", "))
	}
	return nil
}

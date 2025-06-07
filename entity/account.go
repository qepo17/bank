package entity

import (
	"fmt"
	"strings"
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
	Model
}

func (a *Account) Validate() error {
	msgs := []string{}
	if a.Model.ID == 0 {
		msgs = append(msgs, "id is required")
	}
	if len(msgs) > 0 {
		return fmt.Errorf("%w: %s", ErrValidation, strings.Join(msgs, ", "))
	}
	return nil
}

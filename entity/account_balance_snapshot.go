package entity

import (
	"github.com/shopspring/decimal"
)

type AccountBalanceSnapshot struct {
	Model
	AccountID         uint64
	Balance           decimal.Decimal
	LastTransactionID uint64
}

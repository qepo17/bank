package entity

import (
	"database/sql"

	"github.com/shopspring/decimal"
)

type TrxType string

const (
	TrxTypeCredit TrxType = "CREDIT"
	TrxTypeDebit  TrxType = "DEBIT"
)

type Transaction struct {
	Model
	AccountID  uint64
	TransferID sql.NullInt64
	Amount     decimal.Decimal
	TrxType    TrxType
}

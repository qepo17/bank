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

type CreateTransferFundsParams struct {
	SourceAccountID      uint64
	DestinationAccountID uint64
	Amount               decimal.Decimal
}

type CreateTransferFundsResult struct {
	TransferID   uint64
	Success      bool
	ErrorMessage string
}

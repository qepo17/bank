package entity

import "errors"

var (
	ErrDataNotFound      = errors.New("data not found")
	ErrValidation        = errors.New("validation error")
	ErrNoRows            = errors.New("no rows found")
	ErrInsufficientFunds = errors.New("insufficient funds")
)

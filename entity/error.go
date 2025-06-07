package entity

import "errors"

var (
	ErrDataNotFound = errors.New("data not found")
	ErrValidation   = errors.New("validation error")
)

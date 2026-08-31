package domain

import "errors"

var (
	ErrNotFound          = errors.New("resource not found")
	ErrInvalidData       = errors.New("invalid data")
	ErrUnauthorized      = errors.New("unauthorized")
	ErrPaymentDeclined   = errors.New("payment declined")
	ErrWalletUnavailable = errors.New("wallet service unavailable")
	ErrInternalError     = errors.New("internal server error")
)

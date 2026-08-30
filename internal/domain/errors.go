package domain

import "errors"

var (
	ErrNotFound      = errors.New("resource not found")
	ErrInvalidData   = errors.New("invalid data")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrInternalError = errors.New("internal server error")
)

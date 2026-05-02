package usecase

import "errors"

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidInput = errors.New("invalid input")
	ErrEmailExists  = errors.New("email already exists")
)

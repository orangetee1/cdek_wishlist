package domain

import "errors"

var (
	ErrNotFound      = errors.New("not found")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrAlreadyBooked = errors.New("already booked")
)

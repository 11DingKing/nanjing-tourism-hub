package domain

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrInvalidState = errors.New("invalid state transition")
	ErrForbidden    = errors.New("forbidden")
	ErrCapacity     = errors.New("capacity unavailable")
	ErrExpired      = errors.New("expired")
	ErrCancelled    = errors.New("cancelled")
)

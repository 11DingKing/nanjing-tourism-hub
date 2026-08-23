package domain

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound     = errors.New("not found")
	ErrConflict     = errors.New("conflict")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrInvalidState = errors.New("invalid state")
	ErrCapacity     = errors.New("capacity exceeded")
	ErrExpired      = errors.New("expired")
	ErrCanceled     = errors.New("operation canceled")
	ErrDependency   = errors.New("dependency unavailable")
	ErrDuplicate    = errors.New("duplicate request")
)

type Error struct {
	Kind    error
	Op      string
	Entity  string
	ID      string
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	base := e.Op
	if e.Entity != "" {
		base += " " + e.Entity
	}
	if e.ID != "" {
		base += " " + e.ID
	}
	if e.Message != "" {
		base += ": " + e.Message
	}
	if e.Cause != nil {
		base += ": " + e.Cause.Error()
	}
	return base
}

func (e *Error) Unwrap() []error {
	if e.Cause == nil {
		return []error{e.Kind}
	}
	if e.Kind == nil {
		return []error{e.Cause}
	}
	return []error{e.Kind, e.Cause}
}

func Wrap(kind error, op, entity, id, message string, cause error) error {
	return &Error{Kind: kind, Op: op, Entity: entity, ID: id, Message: message, Cause: cause}
}

func IsKind(err, kind error) bool {
	if errors.Is(err, kind) {
		return true
	}
	var target *Error
	return errors.As(err, &target) && errors.Is(target.Kind, kind)
}

func VersionConflict(entity, id string, expected, actual int64) error {
	return Wrap(ErrConflict, "update", entity, id, fmt.Sprintf("version %d does not match %d", expected, actual), nil)
}

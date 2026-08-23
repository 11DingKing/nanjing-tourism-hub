package tourism

import "errors"

var (
	ErrNotFound     = errors.New("tourism object not found")
	ErrConflict     = errors.New("tourism version conflict")
	ErrCapacity     = errors.New("tourism capacity exceeded")
	ErrInvalidState = errors.New("tourism invalid state")
	ErrCanceled     = errors.New("tourism operation canceled")
)

type Error struct {
	Op, Entity, ID string
	Cause          error
}

func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	return e.Op + " " + e.Entity + " " + e.ID
}
func (e *Error) Unwrap() error { return e.Cause }
func wrap(op, entity, id string, cause error) error {
	return &Error{Op: op, Entity: entity, ID: id, Cause: cause}
}

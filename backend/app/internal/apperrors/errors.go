package apperrors

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrForbidden    = errors.New("forbidden")
	ErrUnauthorized = errors.New("unauthorized")
	ErrValidation   = errors.New("validation failed")
	ErrConflict     = errors.New("conflict")
	ErrInternal     = errors.New("internal error")
	ErrGone         = errors.New("gone")
)

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

func NewValidation(msg string) error {
	return &ValidationError{Message: msg}
}

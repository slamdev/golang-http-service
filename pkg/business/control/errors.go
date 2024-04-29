package control

import (
	"errors"
)

type ValidationError struct {
	err string
}

func (e *ValidationError) Error() string {
	return e.err
}

func NewValidationError(err string) *ValidationError {
	return &ValidationError{err: err}
}

func IsValidationError(err error) bool {
	var validationError *ValidationError
	ok := errors.As(err, &validationError)
	return ok
}

type MissingEntityError struct {
	err string
}

func (e *MissingEntityError) Error() string {
	return e.err
}

func NewMissingEntityError(err string) *MissingEntityError {
	return &MissingEntityError{err: err}
}

func IsMissingEntityError(err error) bool {
	var missingEntityError *MissingEntityError
	ok := errors.As(err, &missingEntityError)
	return ok
}

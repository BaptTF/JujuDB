package services

import "fmt"

// ValidationError represents a validation error from a service operation.
// Handlers should return HTTP 400 when receiving this error type.
type ValidationError struct {
	Err error
}

func (e *ValidationError) Error() string {
	return e.Err.Error()
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}

// NewValidationError wraps an error as a ValidationError
func NewValidationError(err error) *ValidationError {
	return &ValidationError{Err: err}
}

// NotFoundError represents a resource not found error.
// Handlers should return HTTP 404 when receiving this error type.
type NotFoundError struct {
	Resource string
	ID       uint
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found (id: %d)", e.Resource, e.ID)
}

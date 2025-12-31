// Package errors provides common error definitions for the metadata system.
package errors

import "errors"

// Common errors that can be used across the metadata system.
var (
	// ErrNotFound is returned when a requested resource is not found.
	ErrNotFound = errors.New("resource not found")

	// ErrAlreadyExists is returned when attempting to create a resource that already exists.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrInvalidInput is returned when the input is invalid.
	ErrInvalidInput = errors.New("invalid input")

	// ErrConnectionFailed is returned when a connection to an external service fails.
	ErrConnectionFailed = errors.New("connection failed")

	// ErrPermissionDenied is returned when the operation is not permitted.
	ErrPermissionDenied = errors.New("permission denied")

	// ErrTimeout is returned when an operation times out.
	ErrTimeout = errors.New("operation timed out")

	// ErrInternal is returned when an internal error occurs.
	ErrInternal = errors.New("internal error")

	// ErrNotImplemented is returned when a feature is not implemented.
	ErrNotImplemented = errors.New("not implemented")
)

// Is reports whether any error in err's chain matches target.
// This is a convenience wrapper around errors.Is.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As finds the first error in err's chain that matches target.
// This is a convenience wrapper around errors.As.
func As(err error, target any) bool {
	return errors.As(err, target)
}

// Wrap wraps an error with additional context.
// Returns nil if err is nil.
func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return &wrappedError{
		msg: message,
		err: err,
	}
}

// wrappedError is an error that wraps another error with additional context.
type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string {
	return e.msg + ": " + e.err.Error()
}

func (e *wrappedError) Unwrap() error {
	return e.err
}

// Package collector provides context utilities for metadata collection operations.
package collector

import (
	"context"
)

// CheckContext checks if the context has been cancelled or deadline exceeded.
// Returns nil if context is still valid, otherwise returns an appropriate error.
func CheckContext(ctx context.Context, source, operation string) error {
	select {
	case <-ctx.Done():
		return WrapContextError(ctx, source, operation)
	default:
		return nil
	}
}

// WrapContextError wraps a context error with the appropriate CollectorError type.
// This should be called when ctx.Err() is known to be non-nil.
func WrapContextError(ctx context.Context, source, operation string) error {
	err := ctx.Err()
	if err == nil {
		return nil
	}

	switch err {
	case context.Canceled:
		return NewCancelledError(source, operation, err)
	case context.DeadlineExceeded:
		return NewDeadlineExceededError(source, operation, err)
	default:
		// For any other context error, treat as cancelled
		return NewCancelledError(source, operation, err)
	}
}

// IsContextError checks if an error is a context cancellation or deadline error.
func IsContextError(err error) bool {
	if err == nil {
		return false
	}
	code := GetErrorCode(err)
	return code == ErrCodeCancelled || code == ErrCodeDeadlineExceeded
}

// IsCancelled checks if an error indicates the operation was cancelled.
func IsCancelled(err error) bool {
	if err == nil {
		return false
	}
	return GetErrorCode(err) == ErrCodeCancelled
}

// IsDeadlineExceeded checks if an error indicates the deadline was exceeded.
func IsDeadlineExceeded(err error) bool {
	if err == nil {
		return false
	}
	return GetErrorCode(err) == ErrCodeDeadlineExceeded
}

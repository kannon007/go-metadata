package collector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestContextCancellationHandling tests that context cancellation is properly handled.
// Feature: metadata-collector, Property 2: Context Cancellation Handling
// **Validates: Requirements 1.6, 8.5**
func TestContextCancellationHandling(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: CheckContext returns appropriate error when context is cancelled
	properties.Property("CheckContext returns CANCELLED error when context is cancelled", prop.ForAll(
		func(source, operation string) bool {
			ctx, cancel := context.WithCancel(context.Background())
			cancel() // Cancel immediately

			err := CheckContext(ctx, source, operation)
			if err == nil {
				t.Log("Expected error but got nil")
				return false
			}

			// Verify error is a CollectorError with CANCELLED code
			var collErr *CollectorError
			if !errors.As(err, &collErr) {
				t.Logf("Expected CollectorError, got %T", err)
				return false
			}

			if collErr.Code != ErrCodeCancelled {
				t.Logf("Expected CANCELLED error code, got %s", collErr.Code)
				return false
			}

			if collErr.Source != source {
				t.Logf("Expected source %q, got %q", source, collErr.Source)
				return false
			}

			if collErr.Operation != operation {
				t.Logf("Expected operation %q, got %q", operation, collErr.Operation)
				return false
			}

			// Verify the underlying cause is context.Canceled
			if !errors.Is(collErr.Cause, context.Canceled) {
				t.Logf("Expected cause to be context.Canceled, got %v", collErr.Cause)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property: CheckContext returns appropriate error when deadline is exceeded
	properties.Property("CheckContext returns DEADLINE_EXCEEDED error when deadline exceeded", prop.ForAll(
		func(source, operation string) bool {
			// Create a context with a very short timeout and wait for it to expire
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()
			
			// Wait for the context to definitely expire
			<-ctx.Done()
			
			// Verify the context is actually expired with DeadlineExceeded
			if ctx.Err() != context.DeadlineExceeded {
				t.Logf("Context error is %v, expected DeadlineExceeded", ctx.Err())
				return false
			}

			err := CheckContext(ctx, source, operation)
			if err == nil {
				t.Log("Expected error but got nil")
				return false
			}

			// Verify error is a CollectorError with DEADLINE_EXCEEDED code
			var collErr *CollectorError
			if !errors.As(err, &collErr) {
				t.Logf("Expected CollectorError, got %T", err)
				return false
			}

			if collErr.Code != ErrCodeDeadlineExceeded {
				t.Logf("Expected DEADLINE_EXCEEDED error code, got %s", collErr.Code)
				return false
			}

			if collErr.Source != source {
				t.Logf("Expected source %q, got %q", source, collErr.Source)
				return false
			}

			if collErr.Operation != operation {
				t.Logf("Expected operation %q, got %q", operation, collErr.Operation)
				return false
			}

			// Verify the underlying cause is context.DeadlineExceeded
			if !errors.Is(collErr.Cause, context.DeadlineExceeded) {
				t.Logf("Expected cause to be context.DeadlineExceeded, got %v", collErr.Cause)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property: CheckContext returns nil when context is valid
	properties.Property("CheckContext returns nil when context is valid", prop.ForAll(
		func(source, operation string) bool {
			ctx := context.Background()

			err := CheckContext(ctx, source, operation)
			if err != nil {
				t.Logf("Expected nil error, got %v", err)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property: IsContextError correctly identifies context errors
	properties.Property("IsContextError correctly identifies context errors", prop.ForAll(
		func(source, operation string) bool {
			// Test cancelled error
			cancelledErr := NewCancelledError(source, operation, context.Canceled)
			if !IsContextError(cancelledErr) {
				t.Log("IsContextError should return true for cancelled error")
				return false
			}

			// Test deadline exceeded error
			deadlineErr := NewDeadlineExceededError(source, operation, context.DeadlineExceeded)
			if !IsContextError(deadlineErr) {
				t.Log("IsContextError should return true for deadline exceeded error")
				return false
			}

			// Test non-context error
			otherErr := NewQueryError(source, operation, errors.New("some error"))
			if IsContextError(otherErr) {
				t.Log("IsContextError should return false for non-context error")
				return false
			}

			// Test nil error
			if IsContextError(nil) {
				t.Log("IsContextError should return false for nil error")
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property: IsCancelled correctly identifies cancelled errors
	properties.Property("IsCancelled correctly identifies cancelled errors", prop.ForAll(
		func(source, operation string) bool {
			cancelledErr := NewCancelledError(source, operation, context.Canceled)
			if !IsCancelled(cancelledErr) {
				t.Log("IsCancelled should return true for cancelled error")
				return false
			}

			deadlineErr := NewDeadlineExceededError(source, operation, context.DeadlineExceeded)
			if IsCancelled(deadlineErr) {
				t.Log("IsCancelled should return false for deadline exceeded error")
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property: IsDeadlineExceeded correctly identifies deadline exceeded errors
	properties.Property("IsDeadlineExceeded correctly identifies deadline exceeded errors", prop.ForAll(
		func(source, operation string) bool {
			deadlineErr := NewDeadlineExceededError(source, operation, context.DeadlineExceeded)
			if !IsDeadlineExceeded(deadlineErr) {
				t.Log("IsDeadlineExceeded should return true for deadline exceeded error")
				return false
			}

			cancelledErr := NewCancelledError(source, operation, context.Canceled)
			if IsDeadlineExceeded(cancelledErr) {
				t.Log("IsDeadlineExceeded should return false for cancelled error")
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

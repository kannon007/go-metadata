// Package retry provides retry logic with exponential backoff for collector operations.
package retry

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"time"

	"go-metadata/internal/collector"
)

// Config defines retry behavior configuration.
type Config struct {
	// MaxRetries is the maximum number of retry attempts (0 means no retries).
	MaxRetries int `json:"max_retries" yaml:"max_retries"`
	// InitialBackoff is the initial wait duration before the first retry.
	InitialBackoff time.Duration `json:"initial_backoff" yaml:"initial_backoff"`
	// MaxBackoff is the maximum wait duration between retries.
	MaxBackoff time.Duration `json:"max_backoff" yaml:"max_backoff"`
	// Multiplier is the factor by which the backoff increases after each retry.
	Multiplier float64 `json:"multiplier" yaml:"multiplier"`
	// Jitter adds randomness to backoff to prevent thundering herd.
	Jitter float64 `json:"jitter" yaml:"jitter"`
}

// DefaultConfig returns the default retry configuration.
var DefaultConfig = Config{
	MaxRetries:     3,
	InitialBackoff: 100 * time.Millisecond,
	MaxBackoff:     10 * time.Second,
	Multiplier:     2.0,
	Jitter:         0.1,
}

// Result contains the outcome of a retry operation.
type Result[T any] struct {
	// Value is the successful result (if any).
	Value T
	// Err is the final error (if operation failed).
	Err error
	// Attempts is the total number of attempts made.
	Attempts int
	// TotalDuration is the total time spent including all retries.
	TotalDuration time.Duration
}

// WithRetry executes the given function with retry logic using exponential backoff.
// It returns immediately on success or permanent errors.
// For transient errors, it retries up to MaxRetries times with exponential backoff.
func WithRetry[T any](ctx context.Context, config Config, fn func(ctx context.Context) (T, error)) Result[T] {
	var zero T
	startTime := time.Now()
	attempts := 0

	for {
		attempts++
		result, err := fn(ctx)

		if err == nil {
			return Result[T]{
				Value:         result,
				Err:           nil,
				Attempts:      attempts,
				TotalDuration: time.Since(startTime),
			}
		}

		// Check if context is cancelled
		if ctx.Err() != nil {
			return Result[T]{
				Value:         zero,
				Err:           ctx.Err(),
				Attempts:      attempts,
				TotalDuration: time.Since(startTime),
			}
		}

		// Check if error is retryable
		if !IsRetryable(err) {
			return Result[T]{
				Value:         zero,
				Err:           err,
				Attempts:      attempts,
				TotalDuration: time.Since(startTime),
			}
		}

		// Check if we've exhausted retries
		if attempts > config.MaxRetries {
			return Result[T]{
				Value:         zero,
				Err:           err,
				Attempts:      attempts,
				TotalDuration: time.Since(startTime),
			}
		}

		// Calculate backoff duration
		backoff := calculateBackoff(config, attempts)

		// Wait for backoff duration or context cancellation
		select {
		case <-ctx.Done():
			return Result[T]{
				Value:         zero,
				Err:           ctx.Err(),
				Attempts:      attempts,
				TotalDuration: time.Since(startTime),
			}
		case <-time.After(backoff):
			// Continue to next retry
		}
	}
}

// Do is a simplified version of WithRetry that returns just the value and error.
func Do[T any](ctx context.Context, config Config, fn func(ctx context.Context) (T, error)) (T, error) {
	result := WithRetry(ctx, config, fn)
	return result.Value, result.Err
}

// DoSimple executes a function that returns only an error with retry logic.
func DoSimple(ctx context.Context, config Config, fn func(ctx context.Context) error) error {
	_, err := Do(ctx, config, func(ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}

// IsRetryable determines if an error should trigger a retry.
// It checks for CollectorError.Retryable flag and common transient errors.
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check if it's a CollectorError with Retryable flag
	var collErr *collector.CollectorError
	if errors.As(err, &collErr) {
		return collErr.Retryable
	}

	// Context errors are not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// By default, unknown errors are not retryable
	return false
}

// calculateBackoff computes the backoff duration for a given attempt.
func calculateBackoff(config Config, attempt int) time.Duration {
	// Calculate base backoff with exponential growth
	backoff := float64(config.InitialBackoff) * math.Pow(config.Multiplier, float64(attempt-1))

	// Apply maximum backoff cap
	if backoff > float64(config.MaxBackoff) {
		backoff = float64(config.MaxBackoff)
	}

	// Apply jitter if configured
	if config.Jitter > 0 {
		jitterRange := backoff * config.Jitter
		jitter := (rand.Float64()*2 - 1) * jitterRange // Random value between -jitterRange and +jitterRange
		backoff += jitter
	}

	// Ensure backoff is not negative
	if backoff < 0 {
		backoff = 0
	}

	return time.Duration(backoff)
}

// CalculateBackoff is exported for testing purposes.
// It computes the backoff duration for a given attempt number.
func CalculateBackoff(config Config, attempt int) time.Duration {
	return calculateBackoff(config, attempt)
}

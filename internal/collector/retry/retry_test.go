package retry

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"go-metadata/internal/collector"
)

// getTestParameters returns the standard test parameters for property tests.
func getTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	return params
}

// genRetryConfig generates a valid RetryConfig for property testing.
func genRetryConfig() gopter.Gen {
	return gopter.CombineGens(
		gen.IntRange(1, 5),                                    // MaxRetries: 1-5
		gen.Int64Range(10, 100),                               // InitialBackoff: 10-100ms
		gen.Int64Range(500, 2000),                             // MaxBackoff: 500-2000ms
		gen.Float64Range(1.5, 3.0),                            // Multiplier: 1.5-3.0
	).Map(func(values []interface{}) Config {
		return Config{
			MaxRetries:     values[0].(int),
			InitialBackoff: time.Duration(values[1].(int64)) * time.Millisecond,
			MaxBackoff:     time.Duration(values[2].(int64)) * time.Millisecond,
			Multiplier:     values[3].(float64),
			Jitter:         0, // Disable jitter for deterministic testing
		}
	})
}

// TestRetryWithExponentialBackoff tests Property 9: Retry with Exponential Backoff
// Feature: metadata-collector, Property 9: Retry with Exponential Backoff
// **Validates: Requirements 8.2, 8.3**
func TestRetryWithExponentialBackoff(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property 9.1: Retry up to MaxRetries times for transient errors
	properties.Property("retries up to MaxRetries times for transient errors", prop.ForAll(
		func(config Config) bool {
			var attempts int32
			ctx := context.Background()

			// Create a function that always returns a retryable error
			fn := func(ctx context.Context) (int, error) {
				atomic.AddInt32(&attempts, 1)
				return 0, collector.NewNetworkError("test", "operation", errors.New("transient"))
			}

			result := WithRetry(ctx, config, fn)

			// Should have attempted MaxRetries + 1 times (initial + retries)
			expectedAttempts := config.MaxRetries + 1
			actualAttempts := int(atomic.LoadInt32(&attempts))

			if actualAttempts != expectedAttempts {
				t.Logf("Expected %d attempts, got %d", expectedAttempts, actualAttempts)
				return false
			}

			if result.Attempts != expectedAttempts {
				t.Logf("Result.Attempts expected %d, got %d", expectedAttempts, result.Attempts)
				return false
			}

			return true
		},
		genRetryConfig(),
	))

	// Property 9.2: Backoff increases exponentially
	properties.Property("backoff increases exponentially up to MaxBackoff", prop.ForAll(
		func(config Config) bool {
			// Test backoff calculation for multiple attempts
			var prevBackoff time.Duration
			for attempt := 1; attempt <= config.MaxRetries+1; attempt++ {
				backoff := CalculateBackoff(config, attempt)

				// Backoff should be positive
				if backoff < 0 {
					t.Logf("Backoff should be non-negative, got %v", backoff)
					return false
				}

				// Backoff should not exceed MaxBackoff
				if backoff > config.MaxBackoff {
					t.Logf("Backoff %v exceeds MaxBackoff %v", backoff, config.MaxBackoff)
					return false
				}

				// For attempts > 1, backoff should generally increase (unless capped)
				if attempt > 1 && backoff < prevBackoff && prevBackoff < config.MaxBackoff {
					t.Logf("Backoff should increase: prev=%v, current=%v", prevBackoff, backoff)
					return false
				}

				prevBackoff = backoff
			}
			return true
		},
		genRetryConfig(),
	))

	// Property 9.3: First backoff is at least InitialBackoff
	properties.Property("first backoff is at least InitialBackoff", prop.ForAll(
		func(config Config) bool {
			backoff := CalculateBackoff(config, 1)
			// With jitter=0, first backoff should equal InitialBackoff
			if backoff != config.InitialBackoff {
				t.Logf("First backoff %v should equal InitialBackoff %v", backoff, config.InitialBackoff)
				return false
			}
			return true
		},
		genRetryConfig(),
	))

	// Property 9.4: Success on first attempt returns immediately
	properties.Property("success on first attempt returns immediately with 1 attempt", prop.ForAll(
		func(config Config) bool {
			var attempts int32
			ctx := context.Background()

			fn := func(ctx context.Context) (string, error) {
				atomic.AddInt32(&attempts, 1)
				return "success", nil
			}

			result := WithRetry(ctx, config, fn)

			if result.Err != nil {
				t.Logf("Expected no error, got %v", result.Err)
				return false
			}

			if result.Value != "success" {
				t.Logf("Expected 'success', got %v", result.Value)
				return false
			}

			if result.Attempts != 1 {
				t.Logf("Expected 1 attempt, got %d", result.Attempts)
				return false
			}

			return true
		},
		genRetryConfig(),
	))

	// Property 9.5: Success after N failures returns correct result
	properties.Property("success after N failures returns correct result", prop.ForAll(
		func(config Config, failuresBeforeSuccess int) bool {
			if failuresBeforeSuccess > config.MaxRetries {
				// Skip if we'd fail before succeeding
				return true
			}

			var attempts int32
			ctx := context.Background()

			fn := func(ctx context.Context) (int, error) {
				current := atomic.AddInt32(&attempts, 1)
				if int(current) <= failuresBeforeSuccess {
					return 0, collector.NewNetworkError("test", "op", errors.New("transient"))
				}
				return 42, nil
			}

			result := WithRetry(ctx, config, fn)

			if result.Err != nil {
				t.Logf("Expected success, got error: %v", result.Err)
				return false
			}

			if result.Value != 42 {
				t.Logf("Expected value 42, got %d", result.Value)
				return false
			}

			expectedAttempts := failuresBeforeSuccess + 1
			if result.Attempts != expectedAttempts {
				t.Logf("Expected %d attempts, got %d", expectedAttempts, result.Attempts)
				return false
			}

			return true
		},
		genRetryConfig(),
		gen.IntRange(0, 3),
	))

	properties.TestingRun(t)
}

// TestPermanentErrorFailFast tests Property 10: Permanent Error Fail-Fast
// Feature: metadata-collector, Property 10: Permanent Error Fail-Fast
// **Validates: Requirements 8.2, 8.3**
func TestPermanentErrorFailFast(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property 10.1: Permanent errors return immediately without retry
	properties.Property("permanent errors return immediately without retry", prop.ForAll(
		func(config Config) bool {
			var attempts int32
			ctx := context.Background()

			// Create a function that returns a non-retryable error
			fn := func(ctx context.Context) (int, error) {
				atomic.AddInt32(&attempts, 1)
				return 0, collector.NewAuthError("test", "operation", errors.New("auth failed"))
			}

			result := WithRetry(ctx, config, fn)

			// Should have attempted exactly once
			actualAttempts := int(atomic.LoadInt32(&attempts))
			if actualAttempts != 1 {
				t.Logf("Expected 1 attempt for permanent error, got %d", actualAttempts)
				return false
			}

			if result.Attempts != 1 {
				t.Logf("Result.Attempts expected 1, got %d", result.Attempts)
				return false
			}

			// Error should be preserved
			if result.Err == nil {
				t.Logf("Expected error, got nil")
				return false
			}

			return true
		},
		genRetryConfig(),
	))

	// Property 10.2: All non-retryable error types fail fast
	properties.Property("all non-retryable error types fail fast", prop.ForAll(
		func(config Config) bool {
			nonRetryableErrors := []*collector.CollectorError{
				collector.NewAuthError("test", "op", nil),
				collector.NewNotFoundError("test", "op", "resource", nil),
				collector.NewInvalidConfigError("test", "field", "reason"),
				collector.NewQueryError("test", "op", nil),
				collector.NewParseError("test", "op", nil),
				collector.NewConnectionClosedError("test", "op"),
				collector.NewPermissionDeniedError("test", "op", nil),
				collector.NewUnsupportedFeatureError("test", "op", "feature"),
			}

			for _, permErr := range nonRetryableErrors {
				var attempts int32
				ctx := context.Background()

				fn := func(ctx context.Context) (int, error) {
					atomic.AddInt32(&attempts, 1)
					return 0, permErr
				}

				result := WithRetry(ctx, config, fn)

				if int(atomic.LoadInt32(&attempts)) != 1 {
					t.Logf("Error %s should fail fast, but got %d attempts", permErr.Code, attempts)
					return false
				}

				if result.Err == nil {
					t.Logf("Expected error for %s", permErr.Code)
					return false
				}
			}

			return true
		},
		genRetryConfig(),
	))

	// Property 10.3: Permanent error preserves original error information
	properties.Property("permanent error preserves original error information", prop.ForAll(
		func(config Config, source, operation string) bool {
			ctx := context.Background()
			originalCause := errors.New("original cause")
			permErr := collector.NewAuthError(source, operation, originalCause)

			fn := func(ctx context.Context) (int, error) {
				return 0, permErr
			}

			result := WithRetry(ctx, config, fn)

			// Check error is preserved
			var collErr *collector.CollectorError
			if !errors.As(result.Err, &collErr) {
				t.Logf("Expected CollectorError, got %T", result.Err)
				return false
			}

			if collErr.Code != collector.ErrCodeAuthError {
				t.Logf("Expected AUTH_ERROR, got %s", collErr.Code)
				return false
			}

			if collErr.Source != source {
				t.Logf("Expected source %s, got %s", source, collErr.Source)
				return false
			}

			if collErr.Operation != operation {
				t.Logf("Expected operation %s, got %s", operation, collErr.Operation)
				return false
			}

			// Check cause is preserved
			unwrapped := errors.Unwrap(collErr)
			if unwrapped == nil || unwrapped.Error() != "original cause" {
				t.Logf("Original cause not preserved")
				return false
			}

			return true
		},
		genRetryConfig(),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	// Property 10.4: Context cancellation is treated as non-retryable
	properties.Property("context cancellation is treated as non-retryable", prop.ForAll(
		func(config Config) bool {
			var attempts int32
			ctx, cancel := context.WithCancel(context.Background())

			fn := func(ctx context.Context) (int, error) {
				current := atomic.AddInt32(&attempts, 1)
				if current == 1 {
					cancel() // Cancel on first attempt
				}
				return 0, ctx.Err()
			}

			result := WithRetry(ctx, config, fn)

			// Should not retry after context cancellation
			if result.Attempts > 2 {
				t.Logf("Should not retry after context cancellation, got %d attempts", result.Attempts)
				return false
			}

			// Error should be context.Canceled
			if !errors.Is(result.Err, context.Canceled) {
				t.Logf("Expected context.Canceled, got %v", result.Err)
				return false
			}

			return true
		},
		genRetryConfig(),
	))

	properties.TestingRun(t)
}

// TestIsRetryable tests the IsRetryable function behavior
func TestIsRetryable(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	properties.Property("retryable errors are correctly identified", prop.ForAll(
		func(source, operation string) bool {
			// Network errors should be retryable
			if !IsRetryable(collector.NewNetworkError(source, operation, nil)) {
				return false
			}

			// Timeout errors should be retryable
			if !IsRetryable(collector.NewTimeoutError(source, operation, nil)) {
				return false
			}

			// Auth errors should not be retryable
			if IsRetryable(collector.NewAuthError(source, operation, nil)) {
				return false
			}

			// nil error should not be retryable
			if IsRetryable(nil) {
				return false
			}

			// Context errors should not be retryable
			if IsRetryable(context.Canceled) {
				return false
			}

			if IsRetryable(context.DeadlineExceeded) {
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 }),
	))

	properties.TestingRun(t)
}

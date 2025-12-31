package collector

import (
	"errors"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// genErrorCode generates a random valid ErrorCode for property testing.
func genErrorCode() gopter.Gen {
	return gen.OneConstOf(
		ErrCodeAuthError,
		ErrCodeNetworkError,
		ErrCodeTimeout,
		ErrCodeNotFound,
		ErrCodeUnsupportedFeature,
		ErrCodeInvalidConfig,
		ErrCodeQueryError,
		ErrCodeParseError,
		ErrCodeConnectionClosed,
		ErrCodePermissionDenied,
		ErrCodeInferenceError,
	)
}

// genSource generates a random source identifier for property testing.
func genSource() gopter.Gen {
	return gen.OneConstOf("mysql", "postgres", "hive", "clickhouse", "oracle")
}

// genCategory generates a random DataSourceCategory for property testing.
func genCategory() gopter.Gen {
	return gen.OneConstOf(
		CategoryRDBMS,
		CategoryDataWarehouse,
		CategoryDocumentDB,
		CategoryKeyValue,
		CategoryMessageQueue,
		CategoryObjectStorage,
	)
}

// genOperation generates a random operation name for property testing.
func genOperation() gopter.Gen {
	return gen.OneConstOf(
		"connect",
		"close",
		"health_check",
		"discover_catalogs",
		"list_schemas",
		"list_tables",
		"fetch_table_metadata",
		"fetch_table_statistics",
		"fetch_partitions",
		"validate_config",
	)
}

// genCollectorError generates a random CollectorError for property testing.
func genCollectorError() gopter.Gen {
	return gopter.CombineGens(
		genErrorCode(),
		gen.AlphaString(),
		genCategory(),
		genSource(),
		genOperation(),
		gen.Bool(), // hasCause
		gen.Bool(), // retryable
	).Map(func(values []interface{}) *CollectorError {
		var cause error
		if values[5].(bool) {
			cause = errors.New("underlying error")
		}
		return &CollectorError{
			Code:      values[0].(ErrorCode),
			Message:   values[1].(string),
			Category:  values[2].(DataSourceCategory),
			Source:    values[3].(string),
			Operation: values[4].(string),
			Cause:     cause,
			Retryable: values[6].(bool),
		}
	})
}


// TestErrorTypingConsistency tests that all CollectorError instances have valid
// error code, source identifier, and operation name.
// Feature: metadata-collector, Property 7: Error Typing Consistency
// **Validates: Requirements 4.6, 5.6, 6.6**
func TestErrorTypingConsistency(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Valid error codes set for validation
	validErrorCodes := map[ErrorCode]bool{
		ErrCodeAuthError:          true,
		ErrCodeNetworkError:       true,
		ErrCodeTimeout:            true,
		ErrCodeNotFound:           true,
		ErrCodeUnsupportedFeature: true,
		ErrCodeInvalidConfig:      true,
		ErrCodeQueryError:         true,
		ErrCodeParseError:         true,
		ErrCodeConnectionClosed:   true,
		ErrCodePermissionDenied:   true,
		ErrCodeInferenceError:     true,
	}

	properties.Property("CollectorError has valid error code, source, and operation", prop.ForAll(
		func(collErr *CollectorError) bool {
			// Check that error code is valid
			if !validErrorCodes[collErr.Code] {
				t.Logf("Invalid error code: %s", collErr.Code)
				return false
			}

			// Check that source is non-empty
			if collErr.Source == "" {
				t.Logf("Source is empty")
				return false
			}

			// Check that operation is non-empty
			if collErr.Operation == "" {
				t.Logf("Operation is empty")
				return false
			}

			// Check that Error() method returns non-empty string
			errStr := collErr.Error()
			if errStr == "" {
				t.Logf("Error() returned empty string")
				return false
			}

			// Check that error string contains the error code
			if !containsString(errStr, string(collErr.Code)) {
				t.Logf("Error string does not contain error code: %s", errStr)
				return false
			}

			// Check that error string contains the category
			if collErr.Category != "" && !containsString(errStr, string(collErr.Category)) {
				t.Logf("Error string does not contain category: %s", errStr)
				return false
			}

			// Check that error string contains the source
			if !containsString(errStr, collErr.Source) {
				t.Logf("Error string does not contain source: %s", errStr)
				return false
			}

			// Check that error string contains the operation
			if !containsString(errStr, collErr.Operation) {
				t.Logf("Error string does not contain operation: %s", errStr)
				return false
			}

			return true
		},
		genCollectorError(),
	))

	properties.Property("Error constructor functions produce valid CollectorError", prop.ForAll(
		func(source, operation string) bool {
			// Test all error constructor functions
			constructors := []struct {
				name     string
				err      *CollectorError
				expected ErrorCode
			}{
				{"NewAuthError", NewAuthError(source, operation, nil), ErrCodeAuthError},
				{"NewNetworkError", NewNetworkError(source, operation, nil), ErrCodeNetworkError},
				{"NewTimeoutError", NewTimeoutError(source, operation, nil), ErrCodeTimeout},
				{"NewNotFoundError", NewNotFoundError(source, operation, "resource", nil), ErrCodeNotFound},
				{"NewUnsupportedFeatureError", NewUnsupportedFeatureError(source, operation, "feature"), ErrCodeUnsupportedFeature},
				{"NewInvalidConfigError", NewInvalidConfigError(source, "field", "reason"), ErrCodeInvalidConfig},
				{"NewQueryError", NewQueryError(source, operation, nil), ErrCodeQueryError},
				{"NewParseError", NewParseError(source, operation, nil), ErrCodeParseError},
				{"NewConnectionClosedError", NewConnectionClosedError(source, operation), ErrCodeConnectionClosed},
				{"NewPermissionDeniedError", NewPermissionDeniedError(source, operation, nil), ErrCodePermissionDenied},
				{"NewInferenceError", NewInferenceError(source, operation, nil), ErrCodeInferenceError},
			}

			for _, tc := range constructors {
				// Check error code matches expected
				if tc.err.Code != tc.expected {
					t.Logf("%s: expected code %s, got %s", tc.name, tc.expected, tc.err.Code)
					return false
				}

				// Check source is set correctly
				if tc.err.Source != source {
					t.Logf("%s: expected source %s, got %s", tc.name, source, tc.err.Source)
					return false
				}

				// Check error implements error interface
				var _ error = tc.err
			}

			return true
		},
		genSource(),
		genOperation(),
	))

	properties.Property("errors.Is works correctly for CollectorError", prop.ForAll(
		func(code ErrorCode, source, operation string) bool {
			err1 := &CollectorError{Code: code, Source: source, Operation: operation}
			err2 := &CollectorError{Code: code, Source: "different", Operation: "different"}
			err3 := &CollectorError{Code: ErrCodeAuthError, Source: source, Operation: operation}

			// Same code should match
			if !errors.Is(err1, err2) {
				t.Logf("errors.Is should return true for same error code")
				return false
			}

			// Different code should not match (unless both are AUTH_ERROR)
			if code != ErrCodeAuthError && errors.Is(err1, err3) {
				t.Logf("errors.Is should return false for different error codes")
				return false
			}

			return true
		},
		genErrorCode(),
		genSource(),
		genOperation(),
	))

	properties.Property("Unwrap returns the cause error", prop.ForAll(
		func(source, operation string, hasCause bool) bool {
			var cause error
			if hasCause {
				cause = errors.New("underlying cause")
			}

			err := NewNetworkError(source, operation, cause)
			unwrapped := errors.Unwrap(err)

			if hasCause {
				if unwrapped == nil {
					t.Logf("Unwrap should return cause when present")
					return false
				}
				if unwrapped.Error() != "underlying cause" {
					t.Logf("Unwrap returned wrong error: %v", unwrapped)
					return false
				}
			} else {
				if unwrapped != nil {
					t.Logf("Unwrap should return nil when no cause")
					return false
				}
			}

			return true
		},
		genSource(),
		genOperation(),
		gen.Bool(),
	))

	properties.Property("IsRetryable correctly identifies retryable errors", prop.ForAll(
		func(source, operation string) bool {
			// Network errors should be retryable
			networkErr := NewNetworkError(source, operation, nil)
			if !IsRetryable(networkErr) {
				t.Logf("Network error should be retryable")
				return false
			}

			// Timeout errors should be retryable
			timeoutErr := NewTimeoutError(source, operation, nil)
			if !IsRetryable(timeoutErr) {
				t.Logf("Timeout error should be retryable")
				return false
			}

			// Auth errors should not be retryable
			authErr := NewAuthError(source, operation, nil)
			if IsRetryable(authErr) {
				t.Logf("Auth error should not be retryable")
				return false
			}

			// NotFound errors should not be retryable
			notFoundErr := NewNotFoundError(source, operation, "resource", nil)
			if IsRetryable(notFoundErr) {
				t.Logf("NotFound error should not be retryable")
				return false
			}

			return true
		},
		genSource(),
		genOperation(),
	))

	properties.Property("GetErrorCode extracts correct error code", prop.ForAll(
		func(collErr *CollectorError) bool {
			code := GetErrorCode(collErr)
			if code != collErr.Code {
				t.Logf("GetErrorCode returned %s, expected %s", code, collErr.Code)
				return false
			}
			return true
		},
		genCollectorError(),
	))

	properties.TestingRun(t)
}

// containsString checks if s contains substr.
func containsString(s, substr string) bool {
	return len(substr) == 0 || (len(s) >= len(substr) && findSubstring(s, substr))
}

// findSubstring checks if s contains substr.
func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

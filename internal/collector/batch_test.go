package collector

import (
	"context"
	"errors"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// mockCollector is a mock implementation of Collector for testing.
type mockCollector struct {
	// failingTables is a set of table names that should fail
	failingTables map[string]bool
	// failError is the error to return for failing tables
	failError error
}

func newMockCollector(failingTables []string, failError error) *mockCollector {
	m := &mockCollector{
		failingTables: make(map[string]bool),
		failError:     failError,
	}
	for _, t := range failingTables {
		m.failingTables[t] = true
	}
	return m
}

func (m *mockCollector) Connect(ctx context.Context) error { return nil }
func (m *mockCollector) Close() error                      { return nil }
func (m *mockCollector) Category() DataSourceCategory      { return CategoryRDBMS }
func (m *mockCollector) Type() string                      { return "mock" }
func (m *mockCollector) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	return &HealthStatus{Connected: true}, nil
}
func (m *mockCollector) DiscoverCatalogs(ctx context.Context) ([]CatalogInfo, error) {
	return []CatalogInfo{{Catalog: "test"}}, nil
}
func (m *mockCollector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	if m.failingTables[catalog] {
		return nil, m.failError
	}
	return []string{"schema1"}, nil
}
func (m *mockCollector) ListTables(ctx context.Context, catalog, schema string, opts *ListOptions) (*TableListResult, error) {
	return &TableListResult{Tables: []string{"table1"}}, nil
}
func (m *mockCollector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*TableMetadata, error) {
	if m.failingTables[table] {
		return nil, m.failError
	}
	return &TableMetadata{
		Catalog: catalog,
		Schema:  schema,
		Name:    table,
		Type:    TableTypeTable,
	}, nil
}
func (m *mockCollector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*TableStatistics, error) {
	if m.failingTables[table] {
		return nil, m.failError
	}
	return &TableStatistics{RowCount: 100}, nil
}
func (m *mockCollector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]PartitionInfo, error) {
	if m.failingTables[table] {
		return nil, m.failError
	}
	return []PartitionInfo{}, nil
}

// TestPartialFailureHandling tests that partial failure handling works correctly.
// Feature: metadata-collector, Property 11: Partial Failure Handling
// **Validates: Requirements 8.4**
func TestPartialFailureHandling(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: Batch operations continue processing after individual failures
	properties.Property("Batch operations continue after individual failures", prop.ForAll(
		func(successTables, failTables []string) bool {
			// Skip if no tables to test
			if len(successTables) == 0 && len(failTables) == 0 {
				return true
			}

			// Create mock collector that fails for specific tables
			failError := NewQueryError("test", "fetch_table_metadata", errors.New("simulated failure"))
			mock := newMockCollector(failTables, failError)
			batch := NewBatchCollector(mock, "test")

			// Combine all tables
			allTables := append(successTables, failTables...)

			// Fetch metadata for all tables
			result := batch.FetchAllTableMetadata(context.Background(), "catalog", "schema", allTables)

			// Verify total count matches input
			if result.TotalCount != len(allTables) {
				t.Logf("TotalCount mismatch: expected %d, got %d", len(allTables), result.TotalCount)
				return false
			}

			// Verify success count matches expected
			if result.SuccessCount != len(successTables) {
				t.Logf("SuccessCount mismatch: expected %d, got %d", len(successTables), result.SuccessCount)
				return false
			}

			// Verify failure count matches expected
			if result.FailureCount != len(failTables) {
				t.Logf("FailureCount mismatch: expected %d, got %d", len(failTables), result.FailureCount)
				return false
			}

			// Verify HasFailures returns correct value
			expectedHasFailures := len(failTables) > 0
			if result.HasFailures() != expectedHasFailures {
				t.Logf("HasFailures mismatch: expected %v, got %v", expectedHasFailures, result.HasFailures())
				return false
			}

			// Verify IsComplete returns correct value
			expectedIsComplete := len(failTables) == 0
			if result.IsComplete() != expectedIsComplete {
				t.Logf("IsComplete mismatch: expected %v, got %v", expectedIsComplete, result.IsComplete())
				return false
			}

			return true
		},
		gen.SliceOfN(5, gen.AlphaString()).Map(func(s []string) []string {
			// Ensure unique table names for success tables
			seen := make(map[string]bool)
			result := make([]string, 0)
			for _, name := range s {
				if name != "" && !seen[name] && !seen["fail_"+name] {
					seen[name] = true
					result = append(result, name)
				}
			}
			return result
		}),
		gen.SliceOfN(3, gen.AlphaString()).Map(func(s []string) []string {
			// Ensure unique table names for fail tables with prefix
			seen := make(map[string]bool)
			result := make([]string, 0)
			for _, name := range s {
				failName := "fail_" + name
				if name != "" && !seen[failName] && !seen[name] {
					seen[failName] = true
					result = append(result, failName)
				}
			}
			return result
		}),
	))

	// Property: PartialResult correctly tracks success and failure counts
	properties.Property("PartialResult correctly tracks counts", prop.ForAll(
		func(numSuccess, numFailure int) bool {
			result := NewPartialResult[string]()

			// Add successes
			for i := 0; i < numSuccess; i++ {
				result.AddResult("success")
			}

			// Add failures
			for i := 0; i < numFailure; i++ {
				result.AddFailure("item", errors.New("error"))
			}

			// Verify counts
			if result.SuccessCount != numSuccess {
				t.Logf("SuccessCount mismatch: expected %d, got %d", numSuccess, result.SuccessCount)
				return false
			}

			if result.FailureCount != numFailure {
				t.Logf("FailureCount mismatch: expected %d, got %d", numFailure, result.FailureCount)
				return false
			}

			if result.TotalCount != numSuccess+numFailure {
				t.Logf("TotalCount mismatch: expected %d, got %d", numSuccess+numFailure, result.TotalCount)
				return false
			}

			if len(result.Results) != numSuccess {
				t.Logf("Results length mismatch: expected %d, got %d", numSuccess, len(result.Results))
				return false
			}

			if len(result.Failures) != numFailure {
				t.Logf("Failures length mismatch: expected %d, got %d", numFailure, len(result.Failures))
				return false
			}

			return true
		},
		gen.IntRange(0, 10),
		gen.IntRange(0, 10),
	))

	// Property: FailureItem captures error code from CollectorError
	properties.Property("FailureItem captures error code from CollectorError", prop.ForAll(
		func(source, operation string) bool {
			result := NewPartialResult[string]()

			// Add failure with CollectorError
			collErr := NewQueryError(source, operation, errors.New("test error"))
			result.AddFailure("item1", collErr)

			if len(result.Failures) != 1 {
				t.Log("Expected 1 failure")
				return false
			}

			if result.Failures[0].ErrorCode != string(ErrCodeQueryError) {
				t.Logf("Expected error code %s, got %s", ErrCodeQueryError, result.Failures[0].ErrorCode)
				return false
			}

			// Add failure with regular error
			result.AddFailure("item2", errors.New("regular error"))

			if len(result.Failures) != 2 {
				t.Log("Expected 2 failures")
				return false
			}

			if result.Failures[1].ErrorCode != "" {
				t.Logf("Expected empty error code for regular error, got %s", result.Failures[1].ErrorCode)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	// Property: Batch operations handle context cancellation gracefully
	properties.Property("Batch operations handle context cancellation", prop.ForAll(
		func(tables []string) bool {
			if len(tables) == 0 {
				return true
			}

			mock := newMockCollector(nil, nil)
			batch := NewBatchCollector(mock, "test")

			// Create cancelled context
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			// Fetch metadata with cancelled context
			result := batch.FetchAllTableMetadata(ctx, "catalog", "schema", tables)

			// All items should be recorded (either as success or failure)
			if result.TotalCount != len(tables) {
				t.Logf("TotalCount mismatch: expected %d, got %d", len(tables), result.TotalCount)
				return false
			}

			// With cancelled context, all should be failures
			if result.FailureCount != len(tables) {
				t.Logf("FailureCount mismatch: expected %d, got %d", len(tables), result.FailureCount)
				return false
			}

			// Verify all failures have CANCELLED error code
			for _, failure := range result.Failures {
				if failure.ErrorCode != string(ErrCodeCancelled) {
					t.Logf("Expected CANCELLED error code, got %s", failure.ErrorCode)
					return false
				}
			}

			return true
		},
		gen.SliceOfN(5, gen.AlphaString()).Map(func(s []string) []string {
			// Filter empty strings
			result := make([]string, 0)
			for _, name := range s {
				if name != "" {
					result = append(result, name)
				}
			}
			return result
		}),
	))

	properties.TestingRun(t)
}

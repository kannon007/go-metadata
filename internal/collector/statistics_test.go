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

// slowMockCollector is a mock collector that simulates slow statistics collection.
type slowMockCollector struct {
	// delay is the time to wait before returning statistics
	delay time.Duration
	// stats is the statistics to return
	stats *TableStatistics
	// err is the error to return (if any)
	err error
}

func newSlowMockCollector(delay time.Duration, stats *TableStatistics, err error) *slowMockCollector {
	return &slowMockCollector{
		delay: delay,
		stats: stats,
		err:   err,
	}
}

func (m *slowMockCollector) Connect(ctx context.Context) error { return nil }
func (m *slowMockCollector) Close() error                      { return nil }
func (m *slowMockCollector) Category() DataSourceCategory      { return CategoryRDBMS }
func (m *slowMockCollector) Type() string                      { return "slow_mock" }
func (m *slowMockCollector) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	return &HealthStatus{Connected: true}, nil
}
func (m *slowMockCollector) DiscoverCatalogs(ctx context.Context) ([]CatalogInfo, error) {
	return []CatalogInfo{{Catalog: "test"}}, nil
}
func (m *slowMockCollector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	return []string{"schema1"}, nil
}
func (m *slowMockCollector) ListTables(ctx context.Context, catalog, schema string, opts *ListOptions) (*TableListResult, error) {
	return &TableListResult{Tables: []string{"table1"}}, nil
}
func (m *slowMockCollector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*TableMetadata, error) {
	return &TableMetadata{
		Catalog: catalog,
		Schema:  schema,
		Name:    table,
		Type:    TableTypeTable,
	}, nil
}
func (m *slowMockCollector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*TableStatistics, error) {
	// Simulate slow operation
	select {
	case <-time.After(m.delay):
		if m.err != nil {
			return nil, m.err
		}
		return m.stats, nil
	case <-ctx.Done():
		// Return partial result on context cancellation
		return &TableStatistics{
			RowCount:    -1, // Indicate partial
			CollectedAt: time.Now(),
		}, WrapContextError(ctx, "test", "fetch_table_statistics")
	}
}
func (m *slowMockCollector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]PartitionInfo, error) {
	return []PartitionInfo{}, nil
}

// TestStatisticsTimeoutBehavior tests that statistics collection respects timeout.
// Feature: metadata-collector, Property 8: Statistics Timeout Behavior
// **Validates: Requirements 7.5**
func TestStatisticsTimeoutBehavior(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: Statistics collection completes within timeout (with tolerance)
	properties.Property("Statistics collection completes within timeout", prop.ForAll(
		func(timeoutMs int) bool {
			if timeoutMs < 10 {
				timeoutMs = 10 // Minimum timeout
			}
			timeout := time.Duration(timeoutMs) * time.Millisecond

			// Create a collector that takes longer than the timeout
			slowDelay := timeout * 10 // Much longer than timeout
			stats := &TableStatistics{RowCount: 1000, DataSizeBytes: 50000, CollectedAt: time.Now()}
			mock := newSlowMockCollector(slowDelay, stats, nil)

			// Collect with timeout
			start := time.Now()
			result, err := FetchTableStatisticsWithTimeout(
				context.Background(),
				mock,
				"catalog", "schema", "table",
				timeout,
				"test",
			)
			elapsed := time.Since(start)

			// Should not return an error
			if err != nil {
				t.Logf("Unexpected error: %v", err)
				return false
			}

			// Should complete within timeout + reasonable tolerance (50% extra)
			tolerance := timeout / 2
			maxExpected := timeout + tolerance
			if elapsed > maxExpected {
				t.Logf("Elapsed time %v exceeded max expected %v (timeout=%v)", elapsed, maxExpected, timeout)
				return false
			}

			// Result should indicate timeout was reached
			if !result.TimeoutReached {
				t.Log("Expected TimeoutReached to be true")
				return false
			}

			// Result should be marked as partial
			if !result.IsPartial {
				t.Log("Expected IsPartial to be true")
				return false
			}

			// Warning should be set
			if result.Warning == "" {
				t.Log("Expected Warning to be set")
				return false
			}

			return true
		},
		gen.IntRange(50, 200), // Timeout in milliseconds
	))

	// Property: Fast statistics collection returns complete results
	properties.Property("Fast statistics collection returns complete results", prop.ForAll(
		func(rowCount int64, dataSize int64) bool {
			timeout := 500 * time.Millisecond

			// Create a collector that completes quickly
			stats := &TableStatistics{
				RowCount:      rowCount,
				DataSizeBytes: dataSize,
				CollectedAt:   time.Now(),
			}
			mock := newSlowMockCollector(10*time.Millisecond, stats, nil)

			// Collect with timeout
			result, err := FetchTableStatisticsWithTimeout(
				context.Background(),
				mock,
				"catalog", "schema", "table",
				timeout,
				"test",
			)

			// Should not return an error
			if err != nil {
				t.Logf("Unexpected error: %v", err)
				return false
			}

			// Result should not indicate timeout
			if result.TimeoutReached {
				t.Log("Expected TimeoutReached to be false")
				return false
			}

			// Result should not be partial
			if result.IsPartial {
				t.Log("Expected IsPartial to be false")
				return false
			}

			// Statistics should match
			if result.Statistics == nil {
				t.Log("Expected Statistics to be non-nil")
				return false
			}

			if result.Statistics.RowCount != rowCount {
				t.Logf("Expected RowCount %d, got %d", rowCount, result.Statistics.RowCount)
				return false
			}

			if result.Statistics.DataSizeBytes != dataSize {
				t.Logf("Expected DataSizeBytes %d, got %d", dataSize, result.Statistics.DataSizeBytes)
				return false
			}

			return true
		},
		gen.Int64Range(0, 1000000),
		gen.Int64Range(0, 1000000000),
	))

	// Property: Zero timeout means no timeout (runs to completion)
	properties.Property("Zero timeout means no timeout", prop.ForAll(
		func(rowCount int64) bool {
			// Create a collector that completes quickly
			stats := &TableStatistics{
				RowCount:    rowCount,
				CollectedAt: time.Now(),
			}
			mock := newSlowMockCollector(10*time.Millisecond, stats, nil)

			// Collect with zero timeout (no timeout)
			result, err := FetchTableStatisticsWithTimeout(
				context.Background(),
				mock,
				"catalog", "schema", "table",
				0, // Zero timeout
				"test",
			)

			// Should not return an error
			if err != nil {
				t.Logf("Unexpected error: %v", err)
				return false
			}

			// Result should not indicate timeout
			if result.TimeoutReached {
				t.Log("Expected TimeoutReached to be false with zero timeout")
				return false
			}

			// Result should not be partial
			if result.IsPartial {
				t.Log("Expected IsPartial to be false with zero timeout")
				return false
			}

			// Statistics should match
			if result.Statistics == nil || result.Statistics.RowCount != rowCount {
				t.Logf("Expected RowCount %d", rowCount)
				return false
			}

			return true
		},
		gen.Int64Range(0, 1000000),
	))

	// Property: Context cancellation is propagated correctly
	properties.Property("Context cancellation is propagated correctly", prop.ForAll(
		func(source string) bool {
			// Create a collector that takes a long time
			stats := &TableStatistics{RowCount: 1000, CollectedAt: time.Now()}
			mock := newSlowMockCollector(10*time.Second, stats, nil)

			// Create a context that will be cancelled
			ctx, cancel := context.WithCancel(context.Background())

			// Cancel after a short delay
			go func() {
				time.Sleep(50 * time.Millisecond)
				cancel()
			}()

			// Collect with long timeout but cancelled context
			_, err := FetchTableStatisticsWithTimeout(
				ctx,
				mock,
				"catalog", "schema", "table",
				5*time.Second, // Long timeout
				source,
			)

			// Should return a cancellation error
			if err == nil {
				t.Log("Expected cancellation error")
				return false
			}

			// Error should be a cancelled error
			if !IsCancelled(err) {
				t.Logf("Expected cancelled error, got: %v", err)
				return false
			}

			return true
		},
		gen.AlphaString(),
	))

	// Property: StatisticsCollector correctly calculates collection duration
	properties.Property("StatisticsCollector correctly calculates collection duration", prop.ForAll(
		func(delayMs int) bool {
			if delayMs < 10 {
				delayMs = 10
			}
			if delayMs > 100 {
				delayMs = 100
			}
			delay := time.Duration(delayMs) * time.Millisecond

			// Create a collector with known delay
			stats := &TableStatistics{RowCount: 100, CollectedAt: time.Now()}
			mock := newSlowMockCollector(delay, stats, nil)

			// Collect with no timeout
			result, err := FetchTableStatisticsWithTimeout(
				context.Background(),
				mock,
				"catalog", "schema", "table",
				0, // No timeout
				"test",
			)

			if err != nil {
				t.Logf("Unexpected error: %v", err)
				return false
			}

			// Collection duration should be at least the delay
			if result.CollectionDuration < delay {
				t.Logf("CollectionDuration %v should be >= delay %v", result.CollectionDuration, delay)
				return false
			}

			// Collection duration should not be too much more than the delay (with generous tolerance for test environment overhead)
			// Allow up to 100% overhead for goroutine scheduling, context switching, etc.
			maxExpected := delay*2 + 100*time.Millisecond
			if result.CollectionDuration > maxExpected {
				t.Logf("CollectionDuration %v exceeded max expected %v (delay=%v)", result.CollectionDuration, maxExpected, delay)
				return false
			}

			return true
		},
		gen.IntRange(10, 100),
	))

	// Property: GetStatisticsTimeout correctly converts seconds to duration
	properties.Property("GetStatisticsTimeout correctly converts seconds to duration", prop.ForAll(
		func(seconds int) bool {
			timeout := GetStatisticsTimeout(seconds)

			if seconds <= 0 {
				// Zero or negative should return zero duration
				if timeout != 0 {
					t.Logf("Expected 0 for seconds=%d, got %v", seconds, timeout)
					return false
				}
			} else {
				// Positive should return correct duration
				expected := time.Duration(seconds) * time.Second
				if timeout != expected {
					t.Logf("Expected %v for seconds=%d, got %v", expected, seconds, timeout)
					return false
				}
			}

			return true
		},
		gen.IntRange(-10, 100),
	))

	// Property: Error from collector is propagated (non-timeout errors)
	properties.Property("Non-timeout errors are propagated correctly", prop.ForAll(
		func(source, operation string) bool {
			// Create a collector that returns an error quickly
			testErr := NewQueryError(source, operation, errors.New("test error"))
			mock := newSlowMockCollector(10*time.Millisecond, nil, testErr)

			// Collect with timeout
			_, err := FetchTableStatisticsWithTimeout(
				context.Background(),
				mock,
				"catalog", "schema", "table",
				500*time.Millisecond,
				source,
			)

			// Should return the error
			if err == nil {
				t.Log("Expected error to be propagated")
				return false
			}

			// Error should be the original error
			var collErr *CollectorError
			if !errors.As(err, &collErr) {
				t.Logf("Expected CollectorError, got %T", err)
				return false
			}

			if collErr.Code != ErrCodeQueryError {
				t.Logf("Expected QUERY_ERROR code, got %s", collErr.Code)
				return false
			}

			return true
		},
		gen.AlphaString(),
		gen.AlphaString(),
	))

	properties.TestingRun(t)
}

// Package collector provides statistics collection utilities with timeout support.
package collector

import (
	"context"
	"time"
)

// StatisticsResult represents the result of a statistics collection operation.
// It can contain partial results if the operation timed out.
type StatisticsResult struct {
	// Statistics contains the collected statistics (may be partial)
	Statistics *TableStatistics `json:"statistics"`
	// IsPartial indicates whether the statistics are partial due to timeout
	IsPartial bool `json:"is_partial"`
	// Warning contains a warning message if the operation was interrupted
	Warning string `json:"warning,omitempty"`
	// CollectionDuration is the time spent collecting statistics
	CollectionDuration time.Duration `json:"collection_duration"`
	// TimeoutReached indicates if the timeout was reached
	TimeoutReached bool `json:"timeout_reached"`
}

// StatisticsCollector provides statistics collection with timeout support.
type StatisticsCollector struct {
	// Timeout is the maximum duration for statistics collection
	Timeout time.Duration
	// Source identifies the collector type (mysql, postgres, hive)
	Source string
}

// NewStatisticsCollector creates a new StatisticsCollector with the given timeout.
func NewStatisticsCollector(timeout time.Duration, source string) *StatisticsCollector {
	return &StatisticsCollector{
		Timeout: timeout,
		Source:  source,
	}
}

// CollectWithTimeout executes a statistics collection function with timeout control.
// If the timeout is reached, it returns partial results collected so far.
// The collectFn should check the context and return partial results when cancelled.
func (sc *StatisticsCollector) CollectWithTimeout(
	ctx context.Context,
	collectFn func(ctx context.Context) (*TableStatistics, error),
) (*StatisticsResult, error) {
	start := time.Now()

	// If no timeout configured, just run the collection
	if sc.Timeout <= 0 {
		stats, err := collectFn(ctx)
		if err != nil {
			return nil, err
		}
		return &StatisticsResult{
			Statistics:         stats,
			IsPartial:          false,
			CollectionDuration: time.Since(start),
			TimeoutReached:     false,
		}, nil
	}

	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, sc.Timeout)
	defer cancel()

	// Channel to receive the result
	type result struct {
		stats *TableStatistics
		err   error
	}
	resultCh := make(chan result, 1)

	// Run collection in goroutine
	go func() {
		stats, err := collectFn(timeoutCtx)
		resultCh <- result{stats: stats, err: err}
	}()

	// Wait for result or timeout or parent context cancellation
	select {
	case res := <-resultCh:
		duration := time.Since(start)
		if res.err != nil {
			// Check if it's a deadline exceeded error (timeout from our timeout context)
			if IsDeadlineExceeded(res.err) {
				// Return partial result with warning
				return &StatisticsResult{
					Statistics:         res.stats,
					IsPartial:          true,
					Warning:            "statistics collection timed out, returning partial results",
					CollectionDuration: duration,
					TimeoutReached:     true,
				}, nil
			}
			// Check if it's a context cancelled error
			if IsCancelled(res.err) {
				return nil, res.err
			}
			return nil, res.err
		}
		return &StatisticsResult{
			Statistics:         res.stats,
			IsPartial:          false,
			CollectionDuration: duration,
			TimeoutReached:     false,
		}, nil

	case <-ctx.Done():
		// Parent context was cancelled - propagate the cancellation error
		return nil, WrapContextError(ctx, sc.Source, "fetch_table_statistics")

	case <-timeoutCtx.Done():
		duration := time.Since(start)
		// Check if parent context was cancelled (takes precedence over timeout)
		if ctx.Err() != nil {
			return nil, WrapContextError(ctx, sc.Source, "fetch_table_statistics")
		}
		// Timeout reached, wait briefly for any partial result
		select {
		case res := <-resultCh:
			// Got a result just after timeout
			return &StatisticsResult{
				Statistics:         res.stats,
				IsPartial:          true,
				Warning:            "statistics collection timed out, returning partial results",
				CollectionDuration: duration,
				TimeoutReached:     true,
			}, nil
		default:
			// No result available, return empty partial result
			return &StatisticsResult{
				Statistics: &TableStatistics{
					CollectedAt: time.Now(),
				},
				IsPartial:          true,
				Warning:            "statistics collection timed out before any data could be collected",
				CollectionDuration: duration,
				TimeoutReached:     true,
			}, nil
		}
	}
}

// FetchTableStatisticsWithTimeout fetches table statistics with timeout control.
// This is a convenience function that wraps a collector's FetchTableStatistics method.
func FetchTableStatisticsWithTimeout(
	ctx context.Context,
	collector Collector,
	catalog, schema, table string,
	timeout time.Duration,
	source string,
) (*StatisticsResult, error) {
	sc := NewStatisticsCollector(timeout, source)
	return sc.CollectWithTimeout(ctx, func(ctx context.Context) (*TableStatistics, error) {
		return collector.FetchTableStatistics(ctx, catalog, schema, table)
	})
}

// GetStatisticsTimeout returns the configured statistics timeout from config.
// Returns 0 if no timeout is configured.
func GetStatisticsTimeout(maxTimeSeconds int) time.Duration {
	if maxTimeSeconds <= 0 {
		return 0
	}
	return time.Duration(maxTimeSeconds) * time.Second
}

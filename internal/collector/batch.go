// Package collector provides batch operation utilities for metadata collection
// with partial failure handling support.
package collector

import (
	"context"
	"fmt"
)

// BatchCollector provides batch operations with partial failure handling.
// It wraps a Collector and provides methods that continue processing
// even when individual items fail.
type BatchCollector struct {
	collector Collector
	source    string
}

// NewBatchCollector creates a new BatchCollector wrapping the given Collector.
func NewBatchCollector(c Collector, source string) *BatchCollector {
	return &BatchCollector{
		collector: c,
		source:    source,
	}
}

// FetchAllTableMetadata fetches metadata for multiple tables with partial failure handling.
// It continues processing remaining tables even if some fail.
func (b *BatchCollector) FetchAllTableMetadata(ctx context.Context, catalog, schema string, tables []string) *PartialResult[*TableMetadata] {
	result := NewPartialResult[*TableMetadata]()

	for _, table := range tables {
		// Check context before each operation
		if err := CheckContext(ctx, b.source, "fetch_all_table_metadata"); err != nil {
			// On context cancellation, add remaining tables as failures and return
			result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
			continue
		}

		metadata, err := b.collector.FetchTableMetadata(ctx, catalog, schema, table)
		if err != nil {
			// Check if it's a context error - if so, we should stop
			if IsContextError(err) {
				result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
				// Continue to add remaining tables as cancelled
				continue
			}
			// For other errors, record failure and continue
			result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
			continue
		}

		result.AddResult(metadata)
	}

	return result
}

// FetchAllTableStatistics fetches statistics for multiple tables with partial failure handling.
func (b *BatchCollector) FetchAllTableStatistics(ctx context.Context, catalog, schema string, tables []string) *PartialResult[*TableStatistics] {
	result := NewPartialResult[*TableStatistics]()

	for _, table := range tables {
		// Check context before each operation
		if err := CheckContext(ctx, b.source, "fetch_all_table_statistics"); err != nil {
			result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
			continue
		}

		stats, err := b.collector.FetchTableStatistics(ctx, catalog, schema, table)
		if err != nil {
			if IsContextError(err) {
				result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
				continue
			}
			result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
			continue
		}

		result.AddResult(stats)
	}

	return result
}

// FetchAllPartitions fetches partition info for multiple tables with partial failure handling.
func (b *BatchCollector) FetchAllPartitions(ctx context.Context, catalog, schema string, tables []string) *PartialResult[[]PartitionInfo] {
	result := NewPartialResult[[]PartitionInfo]()

	for _, table := range tables {
		// Check context before each operation
		if err := CheckContext(ctx, b.source, "fetch_all_partitions"); err != nil {
			result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
			continue
		}

		partitions, err := b.collector.FetchPartitions(ctx, catalog, schema, table)
		if err != nil {
			if IsContextError(err) {
				result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
				continue
			}
			result.AddFailure(fmt.Sprintf("%s.%s.%s", catalog, schema, table), err)
			continue
		}

		result.AddResult(partitions)
	}

	return result
}

// ListAllSchemas lists schemas from multiple catalogs with partial failure handling.
func (b *BatchCollector) ListAllSchemas(ctx context.Context, catalogs []string) *PartialResult[[]string] {
	result := NewPartialResult[[]string]()

	for _, catalog := range catalogs {
		// Check context before each operation
		if err := CheckContext(ctx, b.source, "list_all_schemas"); err != nil {
			result.AddFailure(catalog, err)
			continue
		}

		schemas, err := b.collector.ListSchemas(ctx, catalog)
		if err != nil {
			if IsContextError(err) {
				result.AddFailure(catalog, err)
				continue
			}
			result.AddFailure(catalog, err)
			continue
		}

		result.AddResult(schemas)
	}

	return result
}

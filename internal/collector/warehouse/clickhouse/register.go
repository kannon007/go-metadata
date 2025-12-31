// Package clickhouse provides ClickHouse metadata collector registration.
package clickhouse

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register ClickHouse collector with the factory
	factory.Register(collector.CategoryDataWarehouse, SourceName, func(cfg *config.ConnectorConfig) (collector.Collector, error) {
		return NewCollector(cfg)
	})
}
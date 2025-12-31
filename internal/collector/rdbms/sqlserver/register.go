// Package sqlserver provides SQL Server metadata collector registration.
package sqlserver

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register SQL Server collector with the factory
	factory.Register(collector.CategoryRDBMS, SourceName, func(cfg *config.ConnectorConfig) (collector.Collector, error) {
		return NewCollector(cfg)
	})
}
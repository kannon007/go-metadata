// Package oracle provides Oracle Database metadata collector registration.
package oracle

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register Oracle collector with the factory
	factory.Register(collector.CategoryRDBMS, SourceName, func(cfg *config.ConnectorConfig) (collector.Collector, error) {
		return NewCollector(cfg)
	})
}
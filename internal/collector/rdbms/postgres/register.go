// Package postgres provides a PostgreSQL metadata collector implementation.
package postgres

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register PostgreSQL collector with the default factory
	_ = factory.Register(collector.CategoryRDBMS, SourceName, NewCollector)
}

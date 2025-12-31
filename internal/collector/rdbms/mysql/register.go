// Package mysql provides a MySQL metadata collector implementation.
package mysql

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register MySQL collector with the default factory
	_ = factory.Register(collector.CategoryRDBMS, SourceName, NewCollector)
}

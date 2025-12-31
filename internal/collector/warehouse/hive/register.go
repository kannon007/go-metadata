// Package hive provides a Hive metadata collector implementation.
package hive

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register Hive collector with the default factory
	_ = factory.Register(collector.CategoryDataWarehouse, SourceName, NewCollector)
}

// Package redis provides a Redis metadata collector implementation.
package redis

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register Redis collector with the default factory
	_ = factory.Register(collector.CategoryKeyValue, SourceName, NewCollector)
}
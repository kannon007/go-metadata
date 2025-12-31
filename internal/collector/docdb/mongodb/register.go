// Package mongodb provides a MongoDB metadata collector implementation.
package mongodb

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register MongoDB collector with the default factory
	_ = factory.Register(collector.CategoryDocumentDB, SourceName, NewCollector)
}
// Package elasticsearch provides an Elasticsearch metadata collector implementation.
package elasticsearch

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register Elasticsearch collector with the default factory
	_ = factory.Register(collector.CategoryDocumentDB, SourceName, NewCollector)
}
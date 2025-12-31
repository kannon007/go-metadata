// Package minio provides a MinIO/S3 metadata collector implementation.
package minio

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register MinIO collector with the default factory
	_ = factory.Register(collector.CategoryObjectStorage, SourceName, NewCollector)
}
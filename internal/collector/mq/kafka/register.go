// Package kafka provides a Kafka metadata collector implementation.
package kafka

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register Kafka collector with the default factory
	_ = factory.Register(collector.CategoryMessageQueue, SourceName, NewCollector)
}
// Package rabbitmq provides a RabbitMQ metadata collector implementation.
package rabbitmq

import (
	"go-metadata/internal/collector"
	"go-metadata/internal/collector/factory"
)

func init() {
	// Register RabbitMQ collector with the default factory
	_ = factory.Register(collector.CategoryMessageQueue, SourceName, NewCollector)
}
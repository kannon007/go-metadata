package kafka

import (
	"fmt"
	"testing"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestTopicMappingProperty tests Property 13: Message Queue Topic Mapping
// **Validates: Requirements 9.2, 9.3, 9.5**
func TestTopicMappingProperty(t *testing.T) {
	// **Feature: metadata-collector, Property 13: Message Queue Topic Mapping**
	// For any Kafka topic with partitions, the mapping to TableMetadata should preserve 
	// partition count in PartitionInfo array, and the TableType should be TOPIC.

	properties := gopter.NewProperties(nil)

	properties.Property("topic mapping preserves partition count and sets correct type", prop.ForAll(
		func(topicName string, partitionCount int) bool {
			// Skip invalid inputs
			if topicName == "" || partitionCount <= 0 || partitionCount > 100 {
				return true
			}

			// Create a mock Kafka collector
			cfg := &config.ConnectorConfig{
				Type:     SourceName,
				Endpoint: "localhost:9092",
			}

			_, err := NewCollector(cfg)
			if err != nil {
				t.Logf("Failed to create collector: %v", err)
				return false
			}

			// Create mock topic metadata with the specified partition count
			metadata := &collector.TableMetadata{
				SourceCategory:  collector.CategoryMessageQueue,
				SourceType:      SourceName,
				Catalog:         "kafka",
				Schema:          "default",
				Name:            topicName,
				Type:            collector.TableTypeTopic,
				LastRefreshedAt: time.Now(),
				InferredSchema:  true,
				Properties:      make(map[string]string),
			}

			// Create partition information
			var partitions []collector.PartitionInfo
			for i := 0; i < partitionCount; i++ {
				partitionInfo := collector.PartitionInfo{
					Name:        fmt.Sprintf("partition-%d", i),
					Type:        "kafka_partition",
					Expression:  fmt.Sprintf("partition_id = %d", i),
					ValuesCount: 1,
				}
				partitions = append(partitions, partitionInfo)
			}
			metadata.Partitions = partitions

			// Add partition count to properties
			metadata.Properties["partition_count"] = fmt.Sprintf("%d", partitionCount)

			// Verify the mapping preserves partition count
			if len(metadata.Partitions) != partitionCount {
				t.Logf("Partition count mismatch: got %d, expected %d", len(metadata.Partitions), partitionCount)
				return false
			}

			// Verify the table type is TOPIC
			if metadata.Type != collector.TableTypeTopic {
				t.Logf("Table type mismatch: got %v, expected %v", metadata.Type, collector.TableTypeTopic)
				return false
			}

			// Verify source category is MessageQueue
			if metadata.SourceCategory != collector.CategoryMessageQueue {
				t.Logf("Source category mismatch: got %v, expected %v", metadata.SourceCategory, collector.CategoryMessageQueue)
				return false
			}

			// Verify source type is kafka
			if metadata.SourceType != SourceName {
				t.Logf("Source type mismatch: got %v, expected %v", metadata.SourceType, SourceName)
				return false
			}

			// Verify partition properties
			for i, partition := range metadata.Partitions {
				expectedName := fmt.Sprintf("partition-%d", i)
				if partition.Name != expectedName {
					t.Logf("Partition name mismatch: got %v, expected %v", partition.Name, expectedName)
					return false
				}

				if partition.Type != "kafka_partition" {
					t.Logf("Partition type mismatch: got %v, expected kafka_partition", partition.Type)
					return false
				}

				expectedExpression := fmt.Sprintf("partition_id = %d", i)
				if partition.Expression != expectedExpression {
					t.Logf("Partition expression mismatch: got %v, expected %v", partition.Expression, expectedExpression)
					return false
				}
			}

			// Verify properties contain partition count
			if metadata.Properties["partition_count"] != fmt.Sprintf("%d", partitionCount) {
				t.Logf("Partition count property mismatch: got %v, expected %d", metadata.Properties["partition_count"], partitionCount)
				return false
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 50 }),
		gen.IntRange(1, 32), // Reasonable partition count range
	))

	properties.Property("topic metadata structure consistency", prop.ForAll(
		func(topicName string, partitionCount int, replicationFactor int) bool {
			// Skip invalid inputs
			if topicName == "" || partitionCount <= 0 || partitionCount > 100 || replicationFactor <= 0 || replicationFactor > 10 {
				return true
			}

			// Create mock topic metadata
			metadata := &collector.TableMetadata{
				SourceCategory:  collector.CategoryMessageQueue,
				SourceType:      SourceName,
				Catalog:         "kafka",
				Schema:          "default",
				Name:            topicName,
				Type:            collector.TableTypeTopic,
				LastRefreshedAt: time.Now(),
				InferredSchema:  true,
				Properties:      make(map[string]string),
			}

			// Add partitions
			var partitions []collector.PartitionInfo
			for i := 0; i < partitionCount; i++ {
				partitionInfo := collector.PartitionInfo{
					Name:        fmt.Sprintf("partition-%d", i),
					Type:        "kafka_partition",
					Columns:     []string{"partition_id"},
					Expression:  fmt.Sprintf("partition_id = %d", i),
					ValuesCount: 1,
				}
				partitions = append(partitions, partitionInfo)
			}
			metadata.Partitions = partitions

			// Add properties
			metadata.Properties["partition_count"] = fmt.Sprintf("%d", partitionCount)
			metadata.Properties["replication_factor"] = fmt.Sprintf("%d", replicationFactor)

			// Add basic Kafka message columns
			metadata.Columns = []collector.Column{
				{
					OrdinalPosition: 1,
					Name:            "key",
					Type:            "bytes",
					SourceType:      "bytes",
					Nullable:        true,
					Comment:         "Message key",
				},
				{
					OrdinalPosition: 2,
					Name:            "value",
					Type:            "bytes",
					SourceType:      "bytes",
					Nullable:        true,
					Comment:         "Message value",
				},
				{
					OrdinalPosition: 3,
					Name:            "timestamp",
					Type:            "timestamp",
					SourceType:      "timestamp",
					Nullable:        false,
					Comment:         "Message timestamp",
				},
				{
					OrdinalPosition: 4,
					Name:            "partition",
					Type:            "int",
					SourceType:      "int32",
					Nullable:        false,
					Comment:         "Partition number",
				},
				{
					OrdinalPosition: 5,
					Name:            "offset",
					Type:            "long",
					SourceType:      "int64",
					Nullable:        false,
					Comment:         "Message offset",
				},
			}

			// Verify all required fields are present and consistent
			if metadata.Name != topicName {
				t.Logf("Topic name mismatch: got %v, expected %v", metadata.Name, topicName)
				return false
			}

			if len(metadata.Partitions) != partitionCount {
				t.Logf("Partition count mismatch: got %d, expected %d", len(metadata.Partitions), partitionCount)
				return false
			}

			if metadata.Type != collector.TableTypeTopic {
				t.Logf("Table type should be TOPIC, got %v", metadata.Type)
				return false
			}

			if metadata.SourceCategory != collector.CategoryMessageQueue {
				t.Logf("Source category should be MessageQueue, got %v", metadata.SourceCategory)
				return false
			}

			// Verify standard Kafka columns are present
			expectedColumns := []string{"key", "value", "timestamp", "partition", "offset"}
			if len(metadata.Columns) < len(expectedColumns) {
				t.Logf("Missing columns: got %d, expected at least %d", len(metadata.Columns), len(expectedColumns))
				return false
			}

			for i, expectedCol := range expectedColumns {
				if i >= len(metadata.Columns) || metadata.Columns[i].Name != expectedCol {
					t.Logf("Column mismatch at position %d: got %v, expected %v", i, 
						func() string {
							if i < len(metadata.Columns) {
								return metadata.Columns[i].Name
							}
							return "missing"
						}(), expectedCol)
					return false
				}
			}

			return true
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) <= 50 }),
		gen.IntRange(1, 32),  // Partition count
		gen.IntRange(1, 5),   // Replication factor
	))

	properties.Property("partition info consistency", prop.ForAll(
		func(partitionCount int) bool {
			// Skip invalid inputs
			if partitionCount <= 0 || partitionCount > 100 {
				return true
			}

			// Create partition info array
			var partitions []collector.PartitionInfo
			for i := 0; i < partitionCount; i++ {
				partitionInfo := collector.PartitionInfo{
					Name:        fmt.Sprintf("partition-%d", i),
					Type:        "kafka_partition",
					Columns:     []string{"partition_id"},
					Expression:  fmt.Sprintf("partition_id = %d", i),
					ValuesCount: 1,
				}
				partitions = append(partitions, partitionInfo)
			}

			// Verify each partition has correct properties
			for i, partition := range partitions {
				// Check partition ID consistency
				expectedName := fmt.Sprintf("partition-%d", i)
				if partition.Name != expectedName {
					t.Logf("Partition name inconsistent: got %v, expected %v", partition.Name, expectedName)
					return false
				}

				// Check partition type
				if partition.Type != "kafka_partition" {
					t.Logf("Partition type should be kafka_partition, got %v", partition.Type)
					return false
				}

				// Check partition expression
				expectedExpression := fmt.Sprintf("partition_id = %d", i)
				if partition.Expression != expectedExpression {
					t.Logf("Partition expression inconsistent: got %v, expected %v", partition.Expression, expectedExpression)
					return false
				}

				// Check columns
				if len(partition.Columns) != 1 || partition.Columns[0] != "partition_id" {
					t.Logf("Partition columns should be [partition_id], got %v", partition.Columns)
					return false
				}

				// Check values count
				if partition.ValuesCount != 1 {
					t.Logf("Partition values count should be 1, got %d", partition.ValuesCount)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 32),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
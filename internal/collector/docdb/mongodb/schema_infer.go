// Package mongodb provides schema inference capabilities for MongoDB collections.
package mongodb

import (
	"context"
	"fmt"
	"time"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
	"go-metadata/internal/collector/infer"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// inferSchema infers schema from a MongoDB collection using $sample aggregation
func (c *Collector) inferSchema(ctx context.Context, collection *mongo.Collection) ([]collector.Column, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "infer_schema"); err != nil {
		return nil, err
	}

	// Get sample size from configuration
	sampleSize := c.inferrer.GetConfig().SampleSize
	if sampleSize <= 0 {
		sampleSize = DefaultSampleSize
	}

	// Sample documents using $sample aggregation
	samples, err := c.sampleDocuments(ctx, collection, sampleSize)
	if err != nil {
		return nil, err
	}

	if len(samples) == 0 {
		// Return empty schema for empty collection
		return []collector.Column{}, nil
	}

	// Convert samples to interface{} slice for inferrer
	interfaceSamples := make([]interface{}, len(samples))
	for i, sample := range samples {
		interfaceSamples[i] = sample
	}

	// Use DocumentInferrer to infer schema
	columns, err := c.inferrer.Infer(ctx, interfaceSamples)
	if err != nil {
		return nil, collector.NewInferenceError(SourceName, "infer_schema", err)
	}

	return columns, nil
}

// sampleDocuments samples documents from a MongoDB collection using $sample aggregation
func (c *Collector) sampleDocuments(ctx context.Context, collection *mongo.Collection, sampleSize int) ([]map[string]interface{}, error) {
	// Check context before starting
	if err := collector.CheckContext(ctx, SourceName, "sample_documents"); err != nil {
		return nil, err
	}

	// First, check if collection has any documents
	count, err := collection.EstimatedDocumentCount(ctx)
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "sample_documents")
		}
		return nil, collector.NewQueryError(SourceName, "sample_documents", err)
	}

	if count == 0 {
		return []map[string]interface{}{}, nil
	}

	// Adjust sample size if collection is smaller
	if int64(sampleSize) > count {
		sampleSize = int(count)
	}

	// Build aggregation pipeline with $sample
	pipeline := []bson.D{
		{{"$sample", bson.D{{"size", sampleSize}}}},
	}

	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, pipeline, options.Aggregate().SetMaxTime(time.Duration(*c.getQueryTimeout())*time.Millisecond))
	if err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "sample_documents")
		}
		return nil, collector.NewQueryError(SourceName, "sample_documents", err)
	}
	defer cursor.Close(ctx)

	var samples []map[string]interface{}
	for cursor.Next(ctx) {
		// Check context during iteration
		if err := collector.CheckContext(ctx, SourceName, "sample_documents"); err != nil {
			return nil, err
		}

		var doc map[string]interface{}
		if err := cursor.Decode(&doc); err != nil {
			return nil, collector.NewParseError(SourceName, "sample_documents", err)
		}

		// Remove MongoDB's internal _id field from schema inference if it's ObjectId
		// Keep it if it's a custom type that might be meaningful
		if id, exists := doc["_id"]; exists {
			// Only remove if it's a MongoDB ObjectId (which appears as primitive.ObjectID)
			if fmt.Sprintf("%T", id) == "primitive.ObjectID" {
				delete(doc, "_id")
			}
		}

		samples = append(samples, doc)
	}

	if err := cursor.Err(); err != nil {
		if ctx.Err() != nil {
			return nil, collector.WrapContextError(ctx, SourceName, "sample_documents")
		}
		return nil, collector.NewQueryError(SourceName, "sample_documents", err)
	}

	return samples, nil
}

// getQueryTimeout returns the query timeout duration
func (c *Collector) getQueryTimeout() *int64 {
	if c.config.Properties.ConnectionTimeout > 0 {
		timeout := int64(c.config.Properties.ConnectionTimeout * 1000) // Convert to milliseconds
		return &timeout
	}
	defaultTimeout := int64(DefaultTimeout * 1000)
	return &defaultTimeout
}

// SampleDocuments provides public access to document sampling for testing
// This method is used by the DocumentDBCollector interface
func (c *Collector) SampleDocuments(ctx context.Context, catalog, collection string, limit int) ([]map[string]interface{}, error) {
	if c.client == nil {
		return nil, collector.NewConnectionClosedError(SourceName, "sample_documents")
	}

	db := c.client.Database(catalog)
	coll := db.Collection(collection)

	return c.sampleDocuments(ctx, coll, limit)
}

// SetInferConfig updates the schema inference configuration
// This method is used by the DocumentDBCollector interface
func (c *Collector) SetInferConfig(config *config.InferConfig) {
	if config == nil {
		return
	}

	inferConfig := &infer.InferConfig{
		Enabled:    config.Enabled,
		SampleSize: config.SampleSize,
		MaxDepth:   config.MaxDepth,
		TypeMerge:  infer.TypeMergeStrategy(config.TypeMerge),
	}
	c.inferrer.SetConfig(inferConfig)
}
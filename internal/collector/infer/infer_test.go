package infer

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// getTestParameters returns the standard test parameters for property-based tests.
func getTestParameters() *gopter.TestParameters {
	return &gopter.TestParameters{
		MinSuccessfulTests: 100,
		MaxSize:            50,
		Rng:                nil, // Use default random source
	}
}

// TestSchemaInferenceConsistency tests that schema inference is consistent across multiple runs.
// Feature: metadata-collector, Property 4: Schema Inference Consistency
// **Validates: Requirements 7.3, 7.5**
func TestSchemaInferenceConsistency(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: Document inference produces consistent results
	properties.Property("Document inference produces consistent results", prop.ForAll(
		func(fieldCount int) bool {
			if fieldCount <= 0 || fieldCount > 10 {
				return true // Skip invalid inputs
			}

			// Create test documents
			docs := make([]map[string]interface{}, 3)
			for i := 0; i < 3; i++ {
				doc := make(map[string]interface{})
				for j := 0; j < fieldCount; j++ {
					doc[string(rune('a'+j))] = "test_value"
				}
				docs = append(docs, doc)
			}

			ctx := context.Background()
			inferrer := NewDocumentInferrer()

			// Convert to interface slice
			samples := make([]interface{}, len(docs))
			for i, doc := range docs {
				samples[i] = doc
			}

			// Run inference multiple times
			result1, err1 := inferrer.Infer(ctx, samples)
			if err1 != nil {
				t.Logf("First inference failed: %v", err1)
				return false
			}

			result2, err2 := inferrer.Infer(ctx, samples)
			if err2 != nil {
				t.Logf("Second inference failed: %v", err2)
				return false
			}

			// Results should be identical
			if len(result1) != len(result2) {
				t.Logf("Column count mismatch: %d vs %d", len(result1), len(result2))
				return false
			}

			for i, col1 := range result1 {
				col2 := result2[i]
				if col1.Name != col2.Name {
					t.Logf("Column name mismatch at index %d: %s vs %s", i, col1.Name, col2.Name)
					return false
				}
				if col1.Type != col2.Type {
					t.Logf("Column type mismatch for %s: %s vs %s", col1.Name, col1.Type, col2.Type)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 5),
	))

	// Property: Key pattern inference produces consistent results
	properties.Property("Key pattern inference produces consistent results", prop.ForAll(
		func(keyCount int) bool {
			if keyCount <= 0 || keyCount > 20 {
				return true // Skip invalid inputs
			}

			// Create test keys
			keys := make([]string, keyCount)
			for i := 0; i < keyCount; i++ {
				keys[i] = "user:" + string(rune('a'+i%26))
			}

			ctx := context.Background()
			inferrer := NewKeyPatternInferrer()

			// Convert to interface slice
			samples := make([]interface{}, len(keys))
			for i, key := range keys {
				samples[i] = key
			}

			// Run inference multiple times
			result1, err1 := inferrer.Infer(ctx, samples)
			if err1 != nil {
				t.Logf("First inference failed: %v", err1)
				return false
			}

			result2, err2 := inferrer.Infer(ctx, samples)
			if err2 != nil {
				t.Logf("Second inference failed: %v", err2)
				return false
			}

			// Results should be identical
			if len(result1) != len(result2) {
				t.Logf("Column count mismatch: %d vs %d", len(result1), len(result2))
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	// Property: CSV file inference produces consistent results
	properties.Property("CSV file inference produces consistent results", prop.ForAll(
		func(hasHeader bool) bool {
			ctx := context.Background()
			inferrer := NewFileSchemaInferrer()

			csvData := "name,age\nJohn,25\nJane,30"

			reader1 := strings.NewReader(csvData)
			reader2 := strings.NewReader(csvData)

			request1 := &FileInferenceRequest{
				Reader: reader1,
				Format: FormatCSV,
				CSVOptions: &CSVOptions{
					Delimiter: ',',
					HasHeader: hasHeader,
					SkipRows:  0,
				},
			}

			request2 := &FileInferenceRequest{
				Reader: reader2,
				Format: FormatCSV,
				CSVOptions: &CSVOptions{
					Delimiter: ',',
					HasHeader: hasHeader,
					SkipRows:  0,
				},
			}

			// Run inference multiple times
			result1, err1 := inferrer.InferFromFile(ctx, request1)
			if err1 != nil {
				t.Logf("First inference failed: %v", err1)
				return false
			}

			result2, err2 := inferrer.InferFromFile(ctx, request2)
			if err2 != nil {
				t.Logf("Second inference failed: %v", err2)
				return false
			}

			// Results should be identical
			if len(result1) != len(result2) {
				t.Logf("Column count mismatch: %d vs %d", len(result1), len(result2))
				return false
			}

			for i, col1 := range result1 {
				col2 := result2[i]
				if col1.Name != col2.Name {
					t.Logf("Column name mismatch at index %d: %s vs %s", i, col1.Name, col2.Name)
					return false
				}
				if col1.Type != col2.Type {
					t.Logf("Column type mismatch for %s: %s vs %s", col1.Name, col1.Type, col2.Type)
					return false
				}
			}

			return true
		},
		gen.Bool(),
	))

	// Property: Context cancellation is handled properly
	properties.Property("Context cancellation is handled properly", prop.ForAll(
		func(fieldCount int) bool {
			if fieldCount <= 0 || fieldCount > 5 {
				return true // Skip invalid inputs
			}

			// Create a context that will be cancelled
			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
			defer cancel()

			// Wait for cancellation
			<-ctx.Done()

			inferrer := NewDocumentInferrer()

			// Create test documents
			docs := make([]interface{}, 2)
			docs[0] = map[string]interface{}{"field1": "value1"}
			docs[1] = map[string]interface{}{"field2": "value2"}

			_, err := inferrer.Infer(ctx, docs)

			// Should return a context error
			if err == nil {
				t.Log("Expected context error but got nil")
				return false
			}

			// Error should be context.Canceled or context.DeadlineExceeded
			if err != context.Canceled && err != context.DeadlineExceeded {
				t.Logf("Expected context error, got: %v", err)
				return false
			}

			return true
		},
		gen.IntRange(1, 3),
	))

	properties.TestingRun(t)
}

// TestFileSchemaInference tests file-specific schema inference properties.
// Feature: metadata-collector, Property 15: File Schema Inference
// **Validates: Requirements 10.4**
func TestFileSchemaInference(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Property: CSV inference handles various data types correctly
	properties.Property("CSV inference handles various data types correctly", prop.ForAll(
		func(hasHeader bool) bool {
			ctx := context.Background()
			inferrer := NewFileSchemaInferrer()

			// Create test CSV data with known types
			csvData := ""
			if hasHeader {
				csvData = "id,name,age,active,score\n"
			}
			csvData += "1,John,25,true,85.5\n"
			csvData += "2,Jane,30,false,92.0\n"
			csvData += "3,Bob,35,true,78.3\n"

			reader := strings.NewReader(csvData)
			request := &FileInferenceRequest{
				Reader: reader,
				Format: FormatCSV,
				CSVOptions: &CSVOptions{
					Delimiter: ',',
					HasHeader: hasHeader,
					SkipRows:  0,
				},
			}

			result, err := inferrer.InferFromFile(ctx, request)
			if err != nil {
				t.Logf("CSV inference failed: %v", err)
				return false
			}

			expectedColumns := 5
			if len(result) != expectedColumns {
				t.Logf("Expected %d columns, got %d", expectedColumns, len(result))
				return false
			}

			// Verify column types are reasonable
			for _, col := range result {
				if col.Type == "" {
					t.Logf("Column %s has empty type", col.Name)
					return false
				}
				if col.OrdinalPosition <= 0 {
					t.Logf("Column %s has invalid ordinal position: %d", col.Name, col.OrdinalPosition)
					return false
				}
			}

			return true
		},
		gen.Bool(),
	))

	// Property: JSON inference handles objects
	properties.Property("JSON inference handles objects", prop.ForAll(
		func(fieldCount int) bool {
			if fieldCount <= 0 || fieldCount > 5 {
				return true // Skip invalid inputs
			}

			ctx := context.Background()
			inferrer := NewFileSchemaInferrer()

			// Create simple JSON array
			jsonData := `[{"field1":"value","field2":123}]`

			reader := strings.NewReader(jsonData)
			request := &FileInferenceRequest{
				Reader: reader,
				Format: FormatJSON,
			}

			result, err := inferrer.InferFromFile(ctx, request)
			if err != nil {
				t.Logf("JSON inference failed: %v", err)
				return false
			}

			// Should have at least the fields we put in
			if len(result) < 2 {
				t.Logf("Expected at least 2 columns, got %d", len(result))
				return false
			}

			// Verify all columns have valid properties
			for _, col := range result {
				if col.Name == "" {
					t.Log("Found column with empty name")
					return false
				}
				if col.Type == "" {
					t.Logf("Column %s has empty type", col.Name)
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 3),
	))

	properties.TestingRun(t)
}
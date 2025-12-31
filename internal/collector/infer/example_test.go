package infer

import (
	"context"
	"fmt"
	"strings"
	"testing"
)

// ExampleDocumentInferrer demonstrates how to use the DocumentInferrer.
func ExampleDocumentInferrer() {
	ctx := context.Background()
	inferrer := NewDocumentInferrer()

	// Sample MongoDB-like documents
	documents := []interface{}{
		map[string]interface{}{
			"_id":    "507f1f77bcf86cd799439011",
			"name":   "John Doe",
			"email":  "john@example.com",
			"age":    30,
			"active": true,
			"profile": map[string]interface{}{
				"bio":      "Software developer",
				"location": "San Francisco",
			},
		},
		map[string]interface{}{
			"_id":    "507f1f77bcf86cd799439012",
			"name":   "Jane Smith",
			"email":  "jane@example.com",
			"age":    25,
			"active": false,
			"profile": map[string]interface{}{
				"bio":      "Data scientist",
				"location": "New York",
			},
		},
	}

	columns, err := inferrer.Infer(ctx, documents)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Inferred %d columns:\n", len(columns))
	for _, col := range columns {
		fmt.Printf("- %s: %s (%s)\n", col.Name, col.Type, col.SourceType)
	}

	// Output:
	// Inferred 8 columns:
	// - _id: TEXT (string)
	// - active: BOOLEAN (boolean)
	// - age: BIGINT (integer)
	// - email: TEXT (string)
	// - name: TEXT (string)
	// - profile: JSON (object)
	// - profile.bio: TEXT (string)
	// - profile.location: TEXT (string)
}

// ExampleKeyPatternInferrer demonstrates how to use the KeyPatternInferrer.
func ExampleKeyPatternInferrer() {
	ctx := context.Background()
	inferrer := NewKeyPatternInferrer()

	// Sample Redis keys
	keys := []interface{}{
		"user:123",
		"user:456",
		"user:789",
		"session:abc123:data",
		"session:def456:data",
		"cache:user:123",
		"cache:user:456",
		"counter:visits:2023",
		"counter:visits:2024",
	}

	columns, err := inferrer.Infer(ctx, keys)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Discovered %d key patterns:\n", len(columns))
	for _, col := range columns {
		fmt.Printf("- %s: %s\n", col.Name, col.Comment)
	}

	// Output will show discovered patterns like:
	// - pattern_1: Key pattern: user:* (matches 3 keys)
	// - pattern_2: Key pattern: session:*:data (matches 2 keys)
	// - pattern_3: Key pattern: cache:user:* (matches 2 keys)
	// - pattern_4: Key pattern: counter:visits:* (matches 2 keys)
}

// ExampleFileSchemaInferrer demonstrates how to use the FileSchemaInferrer.
func ExampleFileSchemaInferrer() {
	ctx := context.Background()
	inferrer := NewFileSchemaInferrer()

	// Sample CSV data
	csvData := `id,name,age,salary,active
1,John Doe,30,75000.50,true
2,Jane Smith,25,65000.00,false
3,Bob Johnson,35,85000.75,true`

	reader := strings.NewReader(csvData)
	request := &FileInferenceRequest{
		Reader: reader,
		Format: FormatCSV,
		CSVOptions: &CSVOptions{
			Delimiter: ',',
			HasHeader: true,
			SkipRows:  0,
		},
	}

	columns, err := inferrer.InferFromFile(ctx, request)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Inferred %d columns from CSV:\n", len(columns))
	for _, col := range columns {
		nullable := ""
		if col.Nullable {
			nullable = " (nullable)"
		}
		fmt.Printf("- %s: %s%s\n", col.Name, col.Type, nullable)
	}

	// Output:
	// Inferred 5 columns from CSV:
	// - id: BIGINT
	// - name: TEXT
	// - age: BIGINT
	// - salary: DOUBLE
	// - active: BOOLEAN
}

// TestExamples runs the examples to ensure they work correctly.
func TestExamples(t *testing.T) {
	// Just run the examples to make sure they don't panic
	ExampleDocumentInferrer()
	ExampleKeyPatternInferrer()
	ExampleFileSchemaInferrer()
}
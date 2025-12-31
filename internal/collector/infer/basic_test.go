package infer

import (
	"context"
	"strings"
	"testing"
)

// TestDocumentInferrerBasic tests basic document inference functionality.
func TestDocumentInferrerBasic(t *testing.T) {
	ctx := context.Background()
	inferrer := NewDocumentInferrer()

	// Test with simple documents
	docs := []interface{}{
		map[string]interface{}{
			"name": "John",
			"age":  25,
		},
		map[string]interface{}{
			"name": "Jane",
			"age":  30,
		},
	}

	result, err := inferrer.Infer(ctx, docs)
	if err != nil {
		t.Fatalf("Document inference failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(result))
	}

	// Check that we have name and age columns
	foundName := false
	foundAge := false
	for _, col := range result {
		if col.Name == "name" {
			foundName = true
			if col.Type != "TEXT" {
				t.Errorf("Expected name column to be TEXT, got %s", col.Type)
			}
		}
		if col.Name == "age" {
			foundAge = true
			if col.Type != "BIGINT" {
				t.Errorf("Expected age column to be BIGINT, got %s", col.Type)
			}
		}
	}

	if !foundName {
		t.Error("Name column not found")
	}
	if !foundAge {
		t.Error("Age column not found")
	}
}

// TestKeyPatternInferrerBasic tests basic key pattern inference functionality.
func TestKeyPatternInferrerBasic(t *testing.T) {
	ctx := context.Background()
	inferrer := NewKeyPatternInferrer()

	// Test with Redis-style keys
	keys := []interface{}{
		"user:123",
		"user:456",
		"user:789",
		"session:abc",
		"session:def",
	}

	result, err := inferrer.Infer(ctx, keys)
	if err != nil {
		t.Fatalf("Key pattern inference failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("Expected at least one pattern, got none")
	}

	// Should have at least one column representing patterns
	for _, col := range result {
		if col.Type != "TEXT" {
			t.Errorf("Expected pattern column to be TEXT, got %s", col.Type)
		}
		if col.SourceType != "key_pattern" {
			t.Errorf("Expected source type to be key_pattern, got %s", col.SourceType)
		}
	}
}

// TestFileSchemaInferrerBasic tests basic file schema inference functionality.
func TestFileSchemaInferrerBasic(t *testing.T) {
	ctx := context.Background()
	inferrer := NewFileSchemaInferrer()

	// Test CSV inference
	csvData := "name,age,active\nJohn,25,true\nJane,30,false"
	reader := strings.NewReader(csvData)

	request := &FileInferenceRequest{
		Reader: reader,
		Format: FormatCSV,
		CSVOptions: DefaultCSVOptions(),
	}

	result, err := inferrer.InferFromFile(ctx, request)
	if err != nil {
		t.Fatalf("CSV inference failed: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("Expected 3 columns, got %d", len(result))
	}

	// Check column names and types
	expectedColumns := map[string]string{
		"name":   "TEXT",
		"age":    "BIGINT",
		"active": "BOOLEAN",
	}

	for _, col := range result {
		expectedType, exists := expectedColumns[col.Name]
		if !exists {
			t.Errorf("Unexpected column: %s", col.Name)
			continue
		}
		if col.Type != expectedType {
			t.Errorf("Column %s: expected type %s, got %s", col.Name, expectedType, col.Type)
		}
	}
}

// TestJSONInference tests JSON file inference.
func TestJSONInference(t *testing.T) {
	ctx := context.Background()
	inferrer := NewFileSchemaInferrer()

	// Test JSON array inference
	jsonData := `[{"name":"John","age":25},{"name":"Jane","age":30}]`
	reader := strings.NewReader(jsonData)

	request := &FileInferenceRequest{
		Reader: reader,
		Format: FormatJSON,
	}

	result, err := inferrer.InferFromFile(ctx, request)
	if err != nil {
		t.Fatalf("JSON inference failed: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("Expected 2 columns, got %d", len(result))
	}

	// Check that we have name and age columns
	foundName := false
	foundAge := false
	for _, col := range result {
		if col.Name == "name" {
			foundName = true
		}
		if col.Name == "age" {
			foundAge = true
		}
	}

	if !foundName {
		t.Error("Name column not found")
	}
	if !foundAge {
		t.Error("Age column not found")
	}
}

// TestInferenceConsistency tests that inference produces consistent results.
func TestInferenceConsistency(t *testing.T) {
	ctx := context.Background()
	inferrer := NewDocumentInferrer()

	docs := []interface{}{
		map[string]interface{}{"field1": "value1", "field2": 123},
		map[string]interface{}{"field1": "value2", "field2": 456},
	}

	// Run inference multiple times
	result1, err1 := inferrer.Infer(ctx, docs)
	if err1 != nil {
		t.Fatalf("First inference failed: %v", err1)
	}

	result2, err2 := inferrer.Infer(ctx, docs)
	if err2 != nil {
		t.Fatalf("Second inference failed: %v", err2)
	}

	// Results should be identical
	if len(result1) != len(result2) {
		t.Fatalf("Column count mismatch: %d vs %d", len(result1), len(result2))
	}

	for i, col1 := range result1 {
		col2 := result2[i]
		if col1.Name != col2.Name {
			t.Errorf("Column name mismatch at index %d: %s vs %s", i, col1.Name, col2.Name)
		}
		if col1.Type != col2.Type {
			t.Errorf("Column type mismatch for %s: %s vs %s", col1.Name, col1.Type, col2.Type)
		}
	}
}

// TestContextCancellation tests that context cancellation is handled properly.
func TestContextCancellation(t *testing.T) {
	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	inferrer := NewDocumentInferrer()
	docs := []interface{}{
		map[string]interface{}{"field1": "value1"},
	}

	_, err := inferrer.Infer(ctx, docs)
	if err == nil {
		t.Error("Expected context cancellation error, got nil")
	}

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}
}
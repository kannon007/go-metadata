package tests

import (
	"encoding/json"
	"fmt"
	"go-metadata/internal/lineage"
	"testing"
)

// printLineageResult prints the lineage result in JSON format for debugging.
func printLineageResult(t *testing.T, sql string, result *lineage.LineageResult) {
	t.Helper()
	fmt.Printf("\n=== SQL ===\n%s\n", sql)
	fmt.Printf("=== Lineage Result ===\n")
	if result == nil {
		fmt.Println("nil")
		return
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Logf("Failed to marshal result: %v", err)
		return
	}
	fmt.Println(string(data))
}

// assertColumnCount verifies the number of columns in the result.
func assertColumnCount(t *testing.T, result *lineage.LineageResult, expected int) {
	t.Helper()
	if result == nil {
		t.Fatalf("Result is nil, expected %d columns", expected)
	}
	if len(result.Columns) != expected {
		t.Errorf("Expected %d columns, got %d", expected, len(result.Columns))
	}
}

// assertColumnLineage verifies a specific column lineage.
func assertColumnLineage(t *testing.T, result *lineage.LineageResult, targetCol string, expectedSources []string, expectedOps []string) {
	t.Helper()
	if result == nil {
		t.Fatal("Result is nil")
	}

	var found *lineage.ColumnLineage
	for i := range result.Columns {
		if result.Columns[i].Target.Column == targetCol {
			found = &result.Columns[i]
			break
		}
	}

	if found == nil {
		t.Errorf("Column '%s' not found in result", targetCol)
		return
	}

	// Check sources
	if expectedSources != nil {
		actualSources := make([]string, len(found.Sources))
		for i, src := range found.Sources {
			if src.Table != "" {
				actualSources[i] = src.Table + "." + src.Column
			} else {
				actualSources[i] = src.Column
			}
		}

		if len(actualSources) != len(expectedSources) {
			t.Errorf("Column '%s': expected %d sources, got %d. Expected: %v, Got: %v",
				targetCol, len(expectedSources), len(actualSources), expectedSources, actualSources)
		} else {
			for i, expected := range expectedSources {
				if actualSources[i] != expected {
					t.Errorf("Column '%s': source[%d] expected '%s', got '%s'",
						targetCol, i, expected, actualSources[i])
				}
			}
		}
	}

	// Check operators
	if expectedOps != nil {
		if len(found.Operators) != len(expectedOps) {
			t.Errorf("Column '%s': expected %d operators, got %d. Expected: %v, Got: %v",
				targetCol, len(expectedOps), len(found.Operators), expectedOps, found.Operators)
		} else {
			for i, expected := range expectedOps {
				if found.Operators[i] != expected {
					t.Errorf("Column '%s': operator[%d] expected '%s', got '%s'",
						targetCol, i, expected, found.Operators[i])
				}
			}
		}
	}
}

// assertTargetTable verifies the target table for all columns.
func assertTargetTable(t *testing.T, result *lineage.LineageResult, expectedTable string) {
	t.Helper()
	if result == nil {
		t.Fatal("Result is nil")
	}
	for _, col := range result.Columns {
		if col.Target.Table != expectedTable {
			t.Errorf("Column '%s': expected target table '%s', got '%s'",
				col.Target.Column, expectedTable, col.Target.Table)
		}
	}
}

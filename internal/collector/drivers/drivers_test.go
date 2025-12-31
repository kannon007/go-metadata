package drivers

import (
	"testing"

	"go-metadata/internal/collector/factory"
)

func TestAllCollectorsRegistered(t *testing.T) {
	// This test verifies that importing the drivers package
	// registers all expected collector types with the factory.
	expectedTypes := []string{"mysql", "postgres", "hive"}

	registeredTypesByCategory := factory.ListTypes()

	// Create a map for easy lookup across all categories
	registered := make(map[string]bool)
	for _, types := range registeredTypesByCategory {
		for _, typ := range types {
			registered[typ] = true
		}
	}

	// Check that all expected types are registered
	for _, expected := range expectedTypes {
		if !registered[expected] {
			t.Errorf("Expected collector type %q to be registered, but it was not. Registered types: %v", expected, registeredTypesByCategory)
		}
	}
}

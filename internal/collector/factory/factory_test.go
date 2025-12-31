package factory

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
)

// getTestParameters returns the standard test parameters for property tests.
func getTestParameters() *gopter.TestParameters {
	params := gopter.DefaultTestParameters()
	params.MinSuccessfulTests = 100
	return params
}

// mockCollector is a simple mock implementation of the Collector interface for testing.
type mockCollector struct {
	typeName string
	category collector.DataSourceCategory
}

func (m *mockCollector) Connect(ctx context.Context) error                     { return nil }
func (m *mockCollector) Close() error                                          { return nil }
func (m *mockCollector) Category() collector.DataSourceCategory                { return m.category }
func (m *mockCollector) Type() string                                          { return m.typeName }
func (m *mockCollector) HealthCheck(ctx context.Context) (*collector.HealthStatus, error) {
	return &collector.HealthStatus{Connected: true}, nil
}
func (m *mockCollector) DiscoverCatalogs(ctx context.Context) ([]collector.CatalogInfo, error) {
	return nil, nil
}
func (m *mockCollector) ListSchemas(ctx context.Context, catalog string) ([]string, error) {
	return nil, nil
}
func (m *mockCollector) ListTables(ctx context.Context, catalog, schema string, opts *collector.ListOptions) (*collector.TableListResult, error) {
	return nil, nil
}
func (m *mockCollector) FetchTableMetadata(ctx context.Context, catalog, schema, table string) (*collector.TableMetadata, error) {
	return nil, nil
}
func (m *mockCollector) FetchTableStatistics(ctx context.Context, catalog, schema, table string) (*collector.TableStatistics, error) {
	return nil, nil
}
func (m *mockCollector) FetchPartitions(ctx context.Context, catalog, schema, table string) ([]collector.PartitionInfo, error) {
	return nil, nil
}

// createMockCreator creates a CollectorCreator that returns a mockCollector.
func createMockCreator(typeName string) CollectorCreator {
	return func(cfg *config.ConnectorConfig) (collector.Collector, error) {
		return &mockCollector{typeName: typeName, category: collector.CategoryRDBMS}, nil
	}
}

// createMockCreatorWithCategory creates a CollectorCreator that returns a mockCollector with specified category.
func createMockCreatorWithCategory(typeName string, category collector.DataSourceCategory) CollectorCreator {
	return func(cfg *config.ConnectorConfig) (collector.Collector, error) {
		return &mockCollector{typeName: typeName, category: category}, nil
	}
}


// TestFactoryExtensibility tests Property 1: Factory Extensibility
// *For any* new collector type and its creator function, registering it with the factory
// should succeed, and subsequently creating a collector with that type should return
// a valid instance without modifying existing code.
// Feature: metadata-collector, Property 1: Factory Extensibility
// **Validates: Requirements 1.5, 10.4**
func TestFactoryExtensibility(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Generator for valid type names (non-empty alphanumeric strings)
	genTypeName := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 50
	})

	properties.Property("registering a new type and creating a collector succeeds", prop.ForAll(
		func(typeName string) bool {
			// Create a fresh factory for each test
			factory := NewFactory()

			// Register the type
			creator := createMockCreator(typeName)
			err := factory.Register(collector.CategoryRDBMS, typeName, creator)
			if err != nil {
				t.Logf("Register failed for type %q: %v", typeName, err)
				return false
			}

			// Verify the type is registered
			if !factory.HasType(typeName) {
				t.Logf("Type %q not found after registration", typeName)
				return false
			}

			// Create a collector with valid config
			cfg := &config.ConnectorConfig{
				Type:     typeName,
				Endpoint: "localhost:3306",
			}
			c, err := factory.Create(cfg)
			if err != nil {
				t.Logf("Create failed for type %q: %v", typeName, err)
				return false
			}

			// Verify the collector is not nil
			if c == nil {
				t.Logf("Created collector is nil for type %q", typeName)
				return false
			}

			return true
		},
		genTypeName,
	))

	properties.TestingRun(t)
}


// TestFactoryTypeManagement tests Property 12: Factory Type Management
// *For any* sequence of register and create operations on the factory:
// - Registered types should appear in ListTypes()
// - Creating a registered type should succeed
// - Creating an unregistered type should return a descriptive error
// Feature: metadata-collector, Property 12: Factory Type Management
// **Validates: Requirements 10.2, 10.3, 10.6**
func TestFactoryTypeManagement(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Generator for valid type names (guaranteed non-empty)
	genTypeName := gen.Identifier().Map(func(s string) string {
		if s == "" {
			return "default"
		}
		return s
	})

	// Generator for a list of unique type names (1-5 items)
	genTypeNames := gen.IntRange(1, 5).FlatMap(func(n interface{}) gopter.Gen {
		count := n.(int)
		return gen.SliceOfN(count, genTypeName).Map(func(names []string) []string {
			// Deduplicate and ensure non-empty
			seen := make(map[string]bool)
			unique := make([]string, 0)
			for _, name := range names {
				if !seen[name] && name != "" {
					seen[name] = true
					unique = append(unique, name)
				}
			}
			if len(unique) == 0 {
				return []string{"default"}
			}
			return unique
		})
	}, reflect.TypeOf([]string{}))

	properties.Property("registered types appear in ListTypes", prop.ForAll(
		func(typeNames []string) bool {
			factory := NewFactory()

			// Register all types
			for _, typeName := range typeNames {
				err := factory.Register(collector.CategoryRDBMS, typeName, createMockCreator(typeName))
				if err != nil {
					t.Logf("Register failed for type %q: %v", typeName, err)
					return false
				}
			}

			// Get listed types
			listedTypes := factory.ListAllTypes()

			// Verify all registered types appear in the list
			listedSet := make(map[string]bool)
			for _, lt := range listedTypes {
				listedSet[lt] = true
			}

			for _, typeName := range typeNames {
				if !listedSet[typeName] {
					t.Logf("Type %q not found in ListTypes()", typeName)
					return false
				}
			}

			return true
		},
		genTypeNames,
	))

	properties.Property("creating registered type succeeds", prop.ForAll(
		func(typeName string) bool {
			factory := NewFactory()

			// Register the type
			err := factory.Register(collector.CategoryRDBMS, typeName, createMockCreator(typeName))
			if err != nil {
				return false
			}

			// Create should succeed
			cfg := &config.ConnectorConfig{
				Type:     typeName,
				Endpoint: "localhost:3306",
			}
			c, err := factory.Create(cfg)
			return err == nil && c != nil
		},
		genTypeName,
	))

	// Generator for two different type names
	genTwoDifferentTypes := gopter.CombineGens(
		genTypeName,
		genTypeName,
	).SuchThat(func(values []interface{}) bool {
		return values[0].(string) != values[1].(string)
	}).Map(func(values []interface{}) [2]string {
		return [2]string{values[0].(string), values[1].(string)}
	})

	properties.Property("creating unregistered type returns descriptive error", prop.ForAll(
		func(types [2]string) bool {
			registeredType := types[0]
			unregisteredType := types[1]

			factory := NewFactory()

			// Register one type
			err := factory.Register(collector.CategoryRDBMS, registeredType, createMockCreator(registeredType))
			if err != nil {
				return false
			}

			// Try to create with unregistered type
			cfg := &config.ConnectorConfig{
				Type:     unregisteredType,
				Endpoint: "localhost:3306",
			}
			_, err = factory.Create(cfg)

			// Should fail with descriptive error
			if err == nil {
				t.Logf("Expected error for unregistered type %q", unregisteredType)
				return false
			}

			// Error should be a FactoryError
			var factoryErr *FactoryError
			if !errors.As(err, &factoryErr) {
				t.Logf("Expected FactoryError, got %T", err)
				return false
			}

			// Error message should mention the type
			if factoryErr.Type != unregisteredType {
				t.Logf("Error type mismatch: expected %q, got %q", unregisteredType, factoryErr.Type)
				return false
			}

			return true
		},
		genTwoDifferentTypes,
	))

	properties.TestingRun(t)
}


// TestCategoryRegistrationConsistency tests Property 1: Category Registration Consistency
// *For any* collector type registered with a category, the collector's Category() method
// should return the same category it was registered under.
// Feature: metadata-collector, Property 1: Category Registration Consistency
// **Validates: Requirements 1.3, 2.7**
func TestCategoryRegistrationConsistency(t *testing.T) {
	properties := gopter.NewProperties(getTestParameters())

	// Generator for valid type names (non-empty alphanumeric strings)
	genTypeName := gen.AlphaString().SuchThat(func(s string) bool {
		return len(s) > 0 && len(s) <= 50
	})

	// Generator for valid categories
	genCategory := gen.OneConstOf(
		collector.CategoryRDBMS,
		collector.CategoryDataWarehouse,
		collector.CategoryDocumentDB,
		collector.CategoryKeyValue,
		collector.CategoryMessageQueue,
		collector.CategoryObjectStorage,
	)

	properties.Property("collector Category() method returns the same category it was registered under", prop.ForAll(
		func(typeName string, category collector.DataSourceCategory) bool {
			// Create a fresh factory for each test
			factory := NewFactory()

			// Create a mock creator that returns a collector with the specified category
			creator := createMockCreatorWithCategory(typeName, category)

			// Register the type with the specified category
			err := factory.Register(category, typeName, creator)
			if err != nil {
				t.Logf("Register failed for type %q with category %q: %v", typeName, category, err)
				return false
			}

			// Create a collector with valid config
			cfg := &config.ConnectorConfig{
				Type:     typeName,
				Category: category,
				Endpoint: "localhost:3306",
			}
			c, err := factory.Create(cfg)
			if err != nil {
				t.Logf("Create failed for type %q: %v", typeName, err)
				return false
			}

			// Verify the collector is not nil
			if c == nil {
				t.Logf("Created collector is nil for type %q", typeName)
				return false
			}

			// Verify the collector's Category() method returns the same category
			actualCategory := c.Category()
			if actualCategory != category {
				t.Logf("Category mismatch for type %q: registered with %q, Category() returned %q", 
					typeName, category, actualCategory)
				return false
			}

			return true
		},
		genTypeName,
		genCategory,
	))

	properties.TestingRun(t)
}


// TestRegisterDuplicateType tests that registering a duplicate type fails.
func TestRegisterDuplicateType(t *testing.T) {
	factory := NewFactory()

	// Register first time should succeed
	err := factory.Register(collector.CategoryRDBMS, "mysql", createMockCreator("mysql"))
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Register second time should fail
	err = factory.Register(collector.CategoryRDBMS, "mysql", createMockCreator("mysql"))
	if err == nil {
		t.Fatal("Second registration should fail")
	}

	var factoryErr *FactoryError
	if !errors.As(err, &factoryErr) {
		t.Fatalf("Expected FactoryError, got %T", err)
	}

	if factoryErr.Operation != "register" {
		t.Errorf("Expected operation 'register', got %q", factoryErr.Operation)
	}
}

// TestRegisterEmptyTypeName tests that registering with empty type name fails.
func TestRegisterEmptyTypeName(t *testing.T) {
	factory := NewFactory()

	err := factory.Register(collector.CategoryRDBMS, "", createMockCreator(""))
	if err == nil {
		t.Fatal("Registration with empty type name should fail")
	}

	err = factory.Register(collector.CategoryRDBMS, "   ", createMockCreator(""))
	if err == nil {
		t.Fatal("Registration with whitespace type name should fail")
	}
}

// TestRegisterNilCreator tests that registering with nil creator fails.
func TestRegisterNilCreator(t *testing.T) {
	factory := NewFactory()

	err := factory.Register(collector.CategoryRDBMS, "mysql", nil)
	if err == nil {
		t.Fatal("Registration with nil creator should fail")
	}

	var factoryErr *FactoryError
	if !errors.As(err, &factoryErr) {
		t.Fatalf("Expected FactoryError, got %T", err)
	}
}

// TestCreateWithNilConfig tests that creating with nil config fails.
func TestCreateWithNilConfig(t *testing.T) {
	factory := NewFactory()

	_, err := factory.Create(nil)
	if err == nil {
		t.Fatal("Create with nil config should fail")
	}

	var factoryErr *FactoryError
	if !errors.As(err, &factoryErr) {
		t.Fatalf("Expected FactoryError, got %T", err)
	}
}

// TestCreateWithInvalidConfig tests that creating with invalid config fails.
func TestCreateWithInvalidConfig(t *testing.T) {
	factory := NewFactory()

	// Register a type
	err := factory.Register(collector.CategoryRDBMS, "mysql", createMockCreator("mysql"))
	if err != nil {
		t.Fatalf("Registration should succeed: %v", err)
	}

	// Create with invalid config (missing endpoint)
	cfg := &config.ConnectorConfig{
		Type: "mysql",
		// Endpoint is missing
	}
	_, err = factory.Create(cfg)
	if err == nil {
		t.Fatal("Create with invalid config should fail")
	}

	var factoryErr *FactoryError
	if !errors.As(err, &factoryErr) {
		t.Fatalf("Expected FactoryError, got %T", err)
	}

	if factoryErr.Cause == nil {
		t.Error("FactoryError should have a cause (validation error)")
	}
}

// TestListTypesEmpty tests ListTypes on empty factory.
func TestListTypesEmpty(t *testing.T) {
	factory := NewFactory()

	types := factory.ListAllTypes()
	if len(types) != 0 {
		t.Errorf("Expected empty list, got %v", types)
	}
}

// TestListTypesSorted tests that ListTypes returns sorted results.
func TestListTypesSorted(t *testing.T) {
	factory := NewFactory()

	// Register in non-sorted order
	factory.Register(collector.CategoryRDBMS, "postgres", createMockCreator("postgres"))
	factory.Register(collector.CategoryRDBMS, "mysql", createMockCreator("mysql"))
	factory.Register(collector.CategoryDataWarehouse, "hive", createMockCreator("hive"))

	types := factory.ListAllTypes()
	expected := []string{"hive", "mysql", "postgres"}

	if len(types) != len(expected) {
		t.Fatalf("Expected %d types, got %d", len(expected), len(types))
	}

	for i, typeName := range types {
		if typeName != expected[i] {
			t.Errorf("Expected types[%d] = %q, got %q", i, expected[i], typeName)
		}
	}
}

// TestDefaultFactory tests the global DefaultFactory and convenience functions.
func TestDefaultFactory(t *testing.T) {
	// Note: This test modifies the global DefaultFactory, so it should be
	// careful not to interfere with other tests.

	// Use a unique type name to avoid conflicts
	typeName := "test_default_factory_type"

	// Register using convenience function
	err := Register(collector.CategoryRDBMS, typeName, createMockCreator(typeName))
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Verify it appears in ListAllTypes
	types := ListAllTypes()
	found := false
	for _, tt := range types {
		if tt == typeName {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Type %q not found in ListAllTypes()", typeName)
	}

	// Create using convenience function
	cfg := &config.ConnectorConfig{
		Type:     typeName,
		Endpoint: "localhost:3306",
	}
	c, err := Create(cfg)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if c == nil {
		t.Fatal("Created collector is nil")
	}
}

// TestListTypesByCategory tests the new ListTypes() method that returns types by category.
func TestListTypesByCategory(t *testing.T) {
	factory := NewFactory()

	// Register types in different categories
	factory.Register(collector.CategoryRDBMS, "mysql", createMockCreator("mysql"))
	factory.Register(collector.CategoryRDBMS, "postgres", createMockCreator("postgres"))
	factory.Register(collector.CategoryDataWarehouse, "hive", createMockCreator("hive"))
	factory.Register(collector.CategoryDocumentDB, "mongodb", createMockCreator("mongodb"))

	typesByCategory := factory.ListTypes()

	// Check RDBMS category
	rdbmsTypes := typesByCategory[collector.CategoryRDBMS]
	expectedRDBMS := []string{"mysql", "postgres"}
	if len(rdbmsTypes) != len(expectedRDBMS) {
		t.Errorf("Expected %d RDBMS types, got %d", len(expectedRDBMS), len(rdbmsTypes))
	}
	for i, typeName := range rdbmsTypes {
		if typeName != expectedRDBMS[i] {
			t.Errorf("Expected RDBMS types[%d] = %q, got %q", i, expectedRDBMS[i], typeName)
		}
	}

	// Check DataWarehouse category
	warehouseTypes := typesByCategory[collector.CategoryDataWarehouse]
	if len(warehouseTypes) != 1 || warehouseTypes[0] != "hive" {
		t.Errorf("Expected DataWarehouse types = [\"hive\"], got %v", warehouseTypes)
	}

	// Check DocumentDB category
	docdbTypes := typesByCategory[collector.CategoryDocumentDB]
	if len(docdbTypes) != 1 || docdbTypes[0] != "mongodb" {
		t.Errorf("Expected DocumentDB types = [\"mongodb\"], got %v", docdbTypes)
	}

	// Check empty category
	kvTypes := typesByCategory[collector.CategoryKeyValue]
	if len(kvTypes) != 0 {
		t.Errorf("Expected empty KeyValue types, got %v", kvTypes)
	}
}

// TestListByCategory tests the ListByCategory method.
func TestListByCategory(t *testing.T) {
	factory := NewFactory()

	// Register types in different categories
	factory.Register(collector.CategoryRDBMS, "mysql", createMockCreator("mysql"))
	factory.Register(collector.CategoryRDBMS, "postgres", createMockCreator("postgres"))
	factory.Register(collector.CategoryDataWarehouse, "hive", createMockCreator("hive"))

	// Test RDBMS category
	rdbmsTypes := factory.ListByCategory(collector.CategoryRDBMS)
	expected := []string{"mysql", "postgres"}
	if len(rdbmsTypes) != len(expected) {
		t.Errorf("Expected %d RDBMS types, got %d", len(expected), len(rdbmsTypes))
	}
	for i, typeName := range rdbmsTypes {
		if typeName != expected[i] {
			t.Errorf("Expected RDBMS types[%d] = %q, got %q", i, expected[i], typeName)
		}
	}

	// Test empty category
	kvTypes := factory.ListByCategory(collector.CategoryKeyValue)
	if len(kvTypes) != 0 {
		t.Errorf("Expected empty KeyValue types, got %v", kvTypes)
	}
}

// TestRegisterInvalidCategory tests that registering with invalid category fails.
func TestRegisterInvalidCategory(t *testing.T) {
	factory := NewFactory()

	err := factory.Register("InvalidCategory", "mysql", createMockCreator("mysql"))
	if err == nil {
		t.Fatal("Registration with invalid category should fail")
	}

	var factoryErr *FactoryError
	if !errors.As(err, &factoryErr) {
		t.Fatalf("Expected FactoryError, got %T", err)
	}

	if factoryErr.Operation != "register" {
		t.Errorf("Expected operation 'register', got %q", factoryErr.Operation)
	}
}

// TestRegisterSameTypeInDifferentCategories tests that the same type name can be registered in different categories.
func TestRegisterSameTypeInDifferentCategories(t *testing.T) {
	factory := NewFactory()

	// Register "test" in RDBMS category
	err := factory.Register(collector.CategoryRDBMS, "test", createMockCreator("test"))
	if err != nil {
		t.Fatalf("First registration should succeed: %v", err)
	}

	// Register "test" in DocumentDB category should also succeed
	err = factory.Register(collector.CategoryDocumentDB, "test", createMockCreator("test"))
	if err != nil {
		t.Fatalf("Registration in different category should succeed: %v", err)
	}

	// Verify both are listed
	rdbmsTypes := factory.ListByCategory(collector.CategoryRDBMS)
	if len(rdbmsTypes) != 1 || rdbmsTypes[0] != "test" {
		t.Errorf("Expected RDBMS types = [\"test\"], got %v", rdbmsTypes)
	}

	docdbTypes := factory.ListByCategory(collector.CategoryDocumentDB)
	if len(docdbTypes) != 1 || docdbTypes[0] != "test" {
		t.Errorf("Expected DocumentDB types = [\"test\"], got %v", docdbTypes)
	}
}

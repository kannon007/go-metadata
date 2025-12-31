package factory

import (
	"fmt"
	"sort"
	"strings"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
)

// FactoryError represents an error that occurred during factory operations.
type FactoryError struct {
	Operation string
	Type      string
	Message   string
	Cause     error
}

func (e *FactoryError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("factory %s error for type '%s': %s: %v", e.Operation, e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("factory %s error for type '%s': %s", e.Operation, e.Type, e.Message)
}

func (e *FactoryError) Unwrap() error {
	return e.Cause
}

// CollectorFactory is a factory for creating Collector instances.
// It maintains a registry of collector types and their creator functions.
type CollectorFactory struct {
	reg *registry
}

// NewFactory creates a new CollectorFactory instance.
func NewFactory() *CollectorFactory {
	return &CollectorFactory{
		reg: newRegistry(),
	}
}

// Register registers a collector type with its creator function under a specific category.
// Returns an error if the type is already registered in that category or if the creator is nil.
func (f *CollectorFactory) Register(category collector.DataSourceCategory, typeName string, creator CollectorCreator) error {
	if !collector.IsValidCategory(category) {
		return &FactoryError{
			Operation: "register",
			Type:      typeName,
			Message:   fmt.Sprintf("invalid category: %s", category),
		}
	}

	if strings.TrimSpace(typeName) == "" {
		return &FactoryError{
			Operation: "register",
			Type:      typeName,
			Message:   "type name cannot be empty",
		}
	}

	if creator == nil {
		return &FactoryError{
			Operation: "register",
			Type:      typeName,
			Message:   "creator function cannot be nil",
		}
	}

	if !f.reg.register(category, typeName, creator) {
		return &FactoryError{
			Operation: "register",
			Type:      typeName,
			Message:   fmt.Sprintf("type is already registered in category %s", category),
		}
	}

	return nil
}


// Create creates a new Collector instance based on the provided configuration.
// It validates the configuration before creating the collector.
// Returns an error if the type is not registered or if validation fails.
func (f *CollectorFactory) Create(cfg *config.ConnectorConfig) (collector.Collector, error) {
	if cfg == nil {
		return nil, &FactoryError{
			Operation: "create",
			Type:      "",
			Message:   "configuration cannot be nil",
		}
	}

	// Validate configuration before creating
	if err := cfg.Validate(); err != nil {
		return nil, &FactoryError{
			Operation: "create",
			Type:      cfg.Type,
			Message:   "configuration validation failed",
			Cause:     err,
		}
	}

	// Get the creator for this type
	creator, exists := f.reg.get(cfg.Type)
	if !exists {
		allTypes := f.ListAllTypes()
		var hint string
		if len(allTypes) > 0 {
			hint = fmt.Sprintf("; registered types are: %s", strings.Join(allTypes, ", "))
		} else {
			hint = "; no collector types are registered"
		}
		return nil, &FactoryError{
			Operation: "create",
			Type:      cfg.Type,
			Message:   fmt.Sprintf("unknown collector type%s", hint),
		}
	}

	// Create the collector
	c, err := creator(cfg)
	if err != nil {
		return nil, &FactoryError{
			Operation: "create",
			Type:      cfg.Type,
			Message:   "failed to create collector",
			Cause:     err,
		}
	}

	return c, nil
}

// ListTypes returns a map of all categories to their registered collector types.
func (f *CollectorFactory) ListTypes() map[collector.DataSourceCategory][]string {
	result := f.reg.listAllByCategory()
	// Sort each category's types
	for category, types := range result {
		sort.Strings(types)
		result[category] = types
	}
	return result
}

// ListByCategory returns a sorted list of registered collector types for a specific category.
func (f *CollectorFactory) ListByCategory(category collector.DataSourceCategory) []string {
	types := f.reg.listByCategory(category)
	sort.Strings(types)
	return types
}

// ListAllTypes returns a sorted list of all registered collector types across all categories.
func (f *CollectorFactory) ListAllTypes() []string {
	types := f.reg.list()
	sort.Strings(types)
	return types
}

// HasType checks if a collector type is registered.
func (f *CollectorFactory) HasType(typeName string) bool {
	return f.reg.has(typeName)
}

// DefaultFactory is the global default factory instance.
// Collector implementations should register themselves with this factory
// in their init() functions.
var DefaultFactory = NewFactory()

// Register is a convenience function that registers a collector type
// with the DefaultFactory under a specific category.
func Register(category collector.DataSourceCategory, typeName string, creator CollectorCreator) error {
	return DefaultFactory.Register(category, typeName, creator)
}

// Create is a convenience function that creates a collector using
// the DefaultFactory.
func Create(cfg *config.ConnectorConfig) (collector.Collector, error) {
	return DefaultFactory.Create(cfg)
}

// ListTypes is a convenience function that lists all registered types by category
// from the DefaultFactory.
func ListTypes() map[collector.DataSourceCategory][]string {
	return DefaultFactory.ListTypes()
}

// ListByCategory is a convenience function that lists registered types for a specific category
// from the DefaultFactory.
func ListByCategory(category collector.DataSourceCategory) []string {
	return DefaultFactory.ListByCategory(category)
}

// ListAllTypes is a convenience function that lists all registered types
// from the DefaultFactory.
func ListAllTypes() []string {
	return DefaultFactory.ListAllTypes()
}

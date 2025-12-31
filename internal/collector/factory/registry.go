// Package factory provides a factory pattern implementation for creating
// metadata collectors based on configuration.
package factory

import (
	"sync"

	"go-metadata/internal/collector"
	"go-metadata/internal/collector/config"
)

// CollectorCreator is a function type that creates a Collector instance
// from a ConnectorConfig. Each collector type registers its creator function
// with the factory.
type CollectorCreator func(cfg *config.ConnectorConfig) (collector.Collector, error)

// registry is a thread-safe map that stores registered collector creators organized by category.
// It uses a sync.RWMutex to allow concurrent reads while ensuring exclusive
// access during writes.
type registry struct {
	mu       sync.RWMutex
	creators map[collector.DataSourceCategory]map[string]CollectorCreator
}

// newRegistry creates a new empty registry.
func newRegistry() *registry {
	return &registry{
		creators: make(map[collector.DataSourceCategory]map[string]CollectorCreator),
	}
}

// register adds a collector creator to the registry under the specified category.
// It returns true if the registration was successful (type was not already registered in that category),
// false otherwise.
func (r *registry) register(category collector.DataSourceCategory, typeName string, creator CollectorCreator) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.creators[category] == nil {
		r.creators[category] = make(map[string]CollectorCreator)
	}

	if _, exists := r.creators[category][typeName]; exists {
		return false
	}
	r.creators[category][typeName] = creator
	return true
}

// get retrieves a collector creator from the registry by searching across all categories.
// It returns the creator and true if found, nil and false otherwise.
func (r *registry) get(typeName string) (CollectorCreator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, categoryCreators := range r.creators {
		if creator, exists := categoryCreators[typeName]; exists {
			return creator, true
		}
	}
	return nil, false
}

// getByCategory retrieves a collector creator from the registry for a specific category.
// It returns the creator and true if found, nil and false otherwise.
func (r *registry) getByCategory(category collector.DataSourceCategory, typeName string) (CollectorCreator, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if categoryCreators, exists := r.creators[category]; exists {
		if creator, exists := categoryCreators[typeName]; exists {
			return creator, true
		}
	}
	return nil, false
}

// list returns all registered collector type names across all categories.
func (r *registry) list() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var types []string
	for _, categoryCreators := range r.creators {
		for typeName := range categoryCreators {
			types = append(types, typeName)
		}
	}
	return types
}

// listByCategory returns all registered collector type names for a specific category.
func (r *registry) listByCategory(category collector.DataSourceCategory) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var types []string
	if categoryCreators, exists := r.creators[category]; exists {
		for typeName := range categoryCreators {
			types = append(types, typeName)
		}
	}
	return types
}

// listAllByCategory returns a map of all categories to their registered type names.
func (r *registry) listAllByCategory() map[collector.DataSourceCategory][]string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[collector.DataSourceCategory][]string)
	for category, categoryCreators := range r.creators {
		var types []string
		for typeName := range categoryCreators {
			types = append(types, typeName)
		}
		result[category] = types
	}
	return result
}

// has checks if a collector type is registered in any category.
func (r *registry) has(typeName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, categoryCreators := range r.creators {
		if _, exists := categoryCreators[typeName]; exists {
			return true
		}
	}
	return false
}

// hasByCategory checks if a collector type is registered in a specific category.
func (r *registry) hasByCategory(category collector.DataSourceCategory, typeName string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if categoryCreators, exists := r.creators[category]; exists {
		_, exists := categoryCreators[typeName]
		return exists
	}
	return false
}

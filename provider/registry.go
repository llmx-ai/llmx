package provider

import (
	"fmt"
	"sync"
)

var (
	registry = make(map[string]ProviderFactory)
	mu       sync.RWMutex
)

// Register registers a provider factory
func Register(name string, factory ProviderFactory) {
	mu.Lock()
	defer mu.Unlock()
	registry[name] = factory
}

// New creates a new provider by name
func New(name string, opts map[string]interface{}) (Provider, error) {
	mu.RLock()
	factory, ok := registry[name]
	mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return factory(opts)
}

// List returns all registered provider names
func List() []string {
	mu.RLock()
	defer mu.RUnlock()

	names := make([]string, 0, len(registry))
	for name := range registry {
		names = append(names, name)
	}
	return names
}

// IsRegistered checks if a provider is registered
func IsRegistered(name string) bool {
	mu.RLock()
	defer mu.RUnlock()
	_, ok := registry[name]
	return ok
}

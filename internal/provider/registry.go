package provider

import (
	"fmt"
	"sync"
)

// ProviderFactory creates a new provider instance
type ProviderFactory func(cfg *Config) (Provider, error)

// Registry manages available providers
type Registry struct {
	mu        sync.RWMutex
	factories map[string]ProviderFactory
}

var defaultRegistry = &Registry{
	factories: make(map[string]ProviderFactory),
}

// Register adds a provider factory to the registry
func Register(name string, factory ProviderFactory) {
	defaultRegistry.mu.Lock()
	defer defaultRegistry.mu.Unlock()
	defaultRegistry.factories[name] = factory
}

// GetProvider creates a provider instance by name
func GetProvider(name string, cfg *Config) (Provider, error) {
	defaultRegistry.mu.RLock()
	factory, exists := defaultRegistry.factories[name]
	defaultRegistry.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return factory(cfg)
}

// ListProviders returns all registered provider names
func ListProviders() []string {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()

	names := make([]string, 0, len(defaultRegistry.factories))
	for name := range defaultRegistry.factories {
		names = append(names, name)
	}
	return names
}

// IsRegistered checks if a provider is registered
func IsRegistered(name string) bool {
	defaultRegistry.mu.RLock()
	defer defaultRegistry.mu.RUnlock()
	_, exists := defaultRegistry.factories[name]
	return exists
}

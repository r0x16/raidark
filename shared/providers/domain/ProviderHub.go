package domain

import (
	"fmt"
	"reflect"
)

// ProviderHub is a dependency injection container that stores and retrieves
// providers of different types using reflection and generics.
// It maintains a map of types to their corresponding provider instances.
type ProviderHub struct {
	providers map[reflect.Type]any
}

// Register stores a provider instance in the hub using its type as the key.
// The function uses generics to maintain type safety and automatically
// initializes the providers map if it's nil.
//
// Parameters:
//   - hub: The ProviderHub instance to register the provider in
//   - provider: The provider instance to register
//
// Returns:
//   - The same provider instance for method chaining
func Register[T any](hub *ProviderHub, provider T) T {
	if hub.providers == nil {
		hub.providers = make(map[reflect.Type]any)
	}
	t := reflect.TypeOf((*T)(nil))
	hub.providers[t] = provider
	return provider
}

// Get retrieves a provider instance from the hub by its type.
// The function uses generics to ensure type safety and will panic
// if the requested provider type is not found.
//
// Parameters:
//   - hub: The ProviderHub instance to retrieve the provider from
//
// Returns:
//   - The provider instance of the requested type
//
// Panics:
//   - If the provider type is not found in the hub
func Get[T any](hub *ProviderHub) T {
	t := reflect.TypeOf((*T)(nil))
	provider, ok := hub.providers[t]
	if !ok {
		panic(fmt.Errorf("provider %T not found", t))
	}
	return provider.(T)
}

// Exists checks if a provider of a given type exists in the hub.
//
// Parameters:
//   - hub: The ProviderHub instance to check the provider in
//   - provider: The provider type to check for
//
// Returns:
//   - True if the provider exists, false otherwise
func Exists[T any](hub *ProviderHub) bool {
	t := reflect.TypeOf((*T)(nil))
	_, ok := hub.providers[t]
	return ok
}

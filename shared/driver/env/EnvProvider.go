package env

import (
	"os"
	"strconv"
	"strings"
)

// EnvProvider provides utilities for reading and parsing environment variables
type EnvProvider struct{}

// NewEnvProvider creates a new instance of EnvProvider
func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// GetString gets environment variable as string with default value
func (e *EnvProvider) GetString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetBool gets environment variable as boolean with default value
func (e *EnvProvider) GetBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetInt gets environment variable as integer with default value
func (e *EnvProvider) GetInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetFloat gets environment variable as float64 with default value
func (e *EnvProvider) GetFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseFloat(value, 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// GetSlice gets environment variable as slice (comma-separated) with default value
func (e *EnvProvider) GetSlice(key string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		slice := strings.Split(value, ",")
		for i, v := range slice {
			slice[i] = strings.TrimSpace(v)
		}
		return slice
	}
	return defaultValue
}

// GetSliceWithSeparator gets environment variable as slice with custom separator
func (e *EnvProvider) GetSliceWithSeparator(key, separator string, defaultValue []string) []string {
	if value := os.Getenv(key); value != "" {
		slice := strings.Split(value, separator)
		for i, v := range slice {
			slice[i] = strings.TrimSpace(v)
		}
		return slice
	}
	return defaultValue
}

// IsSet checks if an environment variable is set (not empty)
func (e *EnvProvider) IsSet(key string) bool {
	return os.Getenv(key) != ""
}

// MustGet gets environment variable and panics if not set or empty
func (e *EnvProvider) MustGet(key string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	panic("Environment variable " + key + " is required but not set")
}

// MustGetInt gets environment variable as int and panics if not set, empty, or invalid
func (e *EnvProvider) MustGetInt(key string) int {
	value := e.MustGet(key)
	if parsed, err := strconv.Atoi(value); err == nil {
		return parsed
	}
	panic("Environment variable " + key + " must be a valid integer")
}

// MustGetBool gets environment variable as bool and panics if not set, empty, or invalid
func (e *EnvProvider) MustGetBool(key string) bool {
	value := e.MustGet(key)
	if parsed, err := strconv.ParseBool(value); err == nil {
		return parsed
	}
	panic("Environment variable " + key + " must be a valid boolean")
}

// GetOrDefault is an alias for GetString for backward compatibility
func (e *EnvProvider) GetOrDefault(key, defaultValue string) string {
	return e.GetString(key, defaultValue)
}

// Global instance for convenient access
var DefaultEnvProvider = NewEnvProvider()

// Convenience functions that use the default provider

// GetString gets environment variable as string with default value using default provider
func GetString(key, defaultValue string) string {
	return DefaultEnvProvider.GetString(key, defaultValue)
}

// GetBool gets environment variable as boolean with default value using default provider
func GetBool(key string, defaultValue bool) bool {
	return DefaultEnvProvider.GetBool(key, defaultValue)
}

// GetInt gets environment variable as integer with default value using default provider
func GetInt(key string, defaultValue int) int {
	return DefaultEnvProvider.GetInt(key, defaultValue)
}

// GetFloat gets environment variable as float64 with default value using default provider
func GetFloat(key string, defaultValue float64) float64 {
	return DefaultEnvProvider.GetFloat(key, defaultValue)
}

// GetSlice gets environment variable as slice with default value using default provider
func GetSlice(key string, defaultValue []string) []string {
	return DefaultEnvProvider.GetSlice(key, defaultValue)
}

// IsSet checks if environment variable is set using default provider
func IsSet(key string) bool {
	return DefaultEnvProvider.IsSet(key)
}

// MustGet gets environment variable and panics if not set using default provider
func MustGet(key string) string {
	return DefaultEnvProvider.MustGet(key)
}

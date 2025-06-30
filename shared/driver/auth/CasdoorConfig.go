package auth

import "os"

// CasdoorConfig holds the configuration for Casdoor authentication
type CasdoorConfig struct {
	Endpoint         string
	ClientId         string
	ClientSecret     string
	Certificate      string
	OrganizationName string
	ApplicationName  string
	RedirectURI      string
}

// NewCasdoorConfigFromEnv creates a new CasdoorConfig from environment variables
func NewCasdoorConfigFromEnv() *CasdoorConfig {
	return &CasdoorConfig{
		Endpoint:         getEnvOrDefault("CASDOOR_ENDPOINT", "http://localhost:8000"),
		ClientId:         os.Getenv("CASDOOR_CLIENT_ID"),
		ClientSecret:     os.Getenv("CASDOOR_CLIENT_SECRET"),
		Certificate:      os.Getenv("CASDOOR_CERTIFICATE"),
		OrganizationName: os.Getenv("CASDOOR_ORGANIZATION"),
		ApplicationName:  os.Getenv("CASDOOR_APPLICATION"),
		RedirectURI:      getEnvOrDefault("CASDOOR_REDIRECT_URI", "http://localhost:8080/callback"),
	}
}

// Validate checks if all required configuration fields are present
func (c *CasdoorConfig) Validate() error {
	if c.Endpoint == "" {
		return newCasdoorError("CASDOOR_ENDPOINT is required")
	}
	if c.ClientId == "" {
		return newCasdoorError("CASDOOR_CLIENT_ID is required")
	}
	if c.ClientSecret == "" {
		return newCasdoorError("CASDOOR_CLIENT_SECRET is required")
	}
	if c.Certificate == "" {
		return newCasdoorError("CASDOOR_CERTIFICATE is required")
	}
	if c.OrganizationName == "" {
		return newCasdoorError("CASDOOR_ORGANIZATION is required")
	}
	if c.ApplicationName == "" {
		return newCasdoorError("CASDOOR_APPLICATION is required")
	}
	return nil
}

// getEnvOrDefault returns environment variable value or default if not set
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

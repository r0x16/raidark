package auth

import (
	"fmt"
	"net/url"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"golang.org/x/oauth2"
)

// CasdoorAuthProvider implements the AuthProvider interface using Casdoor
type CasdoorAuthProvider struct {
	config *CasdoorConfig
	client *casdoorsdk.Client
}

// Verify interface implementation
var _ domauth.AuthProvider = &CasdoorAuthProvider{}

// NewCasdoorAuthProvider creates a new CasdoorAuthProvider instance
func NewCasdoorAuthProvider(config *CasdoorConfig) *CasdoorAuthProvider {
	return &CasdoorAuthProvider{
		config: config,
	}
}

// NewCasdoorAuthProviderFromEnv creates a new CasdoorAuthProvider from environment variables
func NewCasdoorAuthProviderFromEnv() *CasdoorAuthProvider {
	config := NewCasdoorConfigFromEnv()
	return NewCasdoorAuthProvider(config)
}

// Initialize the auth provider with configuration
func (c *CasdoorAuthProvider) Initialize() error {
	if err := c.config.Validate(); err != nil {
		return newCasdoorErrorWithCause("failed to validate configuration", err)
	}

	// Initialize the Casdoor client
	c.client = casdoorsdk.NewClient(
		c.config.Endpoint,
		c.config.ClientId,
		c.config.ClientSecret,
		c.config.Certificate,
		c.config.OrganizationName,
		c.config.ApplicationName,
	)

	return nil
}

// GetAuthURL gets OAuth authorization URL for user login
func (c *CasdoorAuthProvider) GetAuthURL(state string) string {
	if c.client == nil {
		return ""
	}

	// Build OAuth authorization URL manually according to Casdoor documentation
	authURL := fmt.Sprintf("%s/login/oauth/authorize", c.config.Endpoint)
	params := url.Values{}
	params.Add("client_id", c.config.ClientId)
	params.Add("redirect_uri", c.config.RedirectURI)
	params.Add("response_type", "code")
	params.Add("scope", "openid profile email")
	params.Add("state", state)

	return fmt.Sprintf("%s?%s", authURL, params.Encode())
}

// GetToken exchanges authorization code for access token
func (c *CasdoorAuthProvider) GetToken(code, state string) (*oauth2.Token, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	token, err := c.client.GetOAuthToken(code, state)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to get OAuth token", err)
	}

	return token, nil
}

// ParseToken parses and validates JWT token
func (c *CasdoorAuthProvider) ParseToken(token string) (*casdoorsdk.Claims, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	claims, err := c.client.ParseJwtToken(token)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to parse JWT token", err)
	}

	return claims, nil
}

// GetUser gets user information by username
func (c *CasdoorAuthProvider) GetUser(username string) (*casdoorsdk.User, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	user, err := c.client.GetUser(username)
	if err != nil {
		return nil, newCasdoorErrorWithCause(fmt.Sprintf("failed to get user: %s", username), err)
	}

	return user, nil
}

// GetUsers gets all users
func (c *CasdoorAuthProvider) GetUsers() ([]*casdoorsdk.User, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	users, err := c.client.GetUsers()
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to get users", err)
	}

	return users, nil
}

// AddUser creates a new user
func (c *CasdoorAuthProvider) AddUser(user *casdoorsdk.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	success, err := c.client.AddUser(user)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to add user", err)
	}

	return success, nil
}

// UpdateUser updates existing user
func (c *CasdoorAuthProvider) UpdateUser(user *casdoorsdk.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	success, err := c.client.UpdateUser(user)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to update user", err)
	}

	return success, nil
}

// DeleteUser deletes user
func (c *CasdoorAuthProvider) DeleteUser(user *casdoorsdk.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	success, err := c.client.DeleteUser(user)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to delete user", err)
	}

	return success, nil
}

// HealthCheck verifies if provider is healthy
func (c *CasdoorAuthProvider) HealthCheck() error {
	if c.client == nil {
		return newCasdoorError("client not initialized")
	}

	// Try to get users as a simple health check
	_, err := c.client.GetUsers()
	if err != nil {
		return newCasdoorErrorWithCause("health check failed", err)
	}

	return nil
}

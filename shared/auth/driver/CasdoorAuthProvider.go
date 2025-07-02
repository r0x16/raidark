package driver

import (
	"fmt"
	"net/url"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/r0x16/Raidark/shared/auth/domain"
	"golang.org/x/oauth2"
)

// CasdoorAuthProvider implements the AuthProvider interface using Casdoor
type CasdoorAuthProvider struct {
	config *CasdoorConfig
	client *casdoorsdk.Client
}

// Verify interface implementation
var _ domain.AuthProvider = &CasdoorAuthProvider{}

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
func (c *CasdoorAuthProvider) GetToken(code, state string) (*domain.Token, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	oauthToken, err := c.client.GetOAuthToken(code, state)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to get OAuth token", err)
	}

	return c.convertOAuth2TokenToDomainToken(oauthToken), nil
}

// RefreshToken refreshes OAuth token using refresh token
func (c *CasdoorAuthProvider) RefreshToken(refreshToken string) (*domain.Token, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	oauthToken, err := c.client.RefreshOAuthToken(refreshToken)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to refresh OAuth token", err)
	}

	return c.convertOAuth2TokenToDomainToken(oauthToken), nil
}

// ParseToken parses and validates JWT token
func (c *CasdoorAuthProvider) ParseToken(token string) (*domain.Claims, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	casdoorClaims, err := c.client.ParseJwtToken(token)
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to parse JWT token", err)
	}

	return c.convertCasdoorClaimsToDomainClaims(casdoorClaims), nil
}

// GetUser gets user information by username
func (c *CasdoorAuthProvider) GetUser(username string) (*domain.User, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	casdoorUser, err := c.client.GetUser(username)
	if err != nil {
		return nil, newCasdoorErrorWithCause(fmt.Sprintf("failed to get user: %s", username), err)
	}

	// Pure composition - direct embedding
	return &domain.User{User: *casdoorUser}, nil
}

// GetUsers gets all users
func (c *CasdoorAuthProvider) GetUsers() ([]*domain.User, error) {
	if c.client == nil {
		return nil, newCasdoorError("client not initialized")
	}

	casdoorUsers, err := c.client.GetUsers()
	if err != nil {
		return nil, newCasdoorErrorWithCause("failed to get users", err)
	}

	// Pure composition conversion
	domainUsers := make([]*domain.User, len(casdoorUsers))
	for i, casdoorUser := range casdoorUsers {
		domainUsers[i] = &domain.User{User: *casdoorUser}
	}

	return domainUsers, nil
}

// AddUser creates a new user
func (c *CasdoorAuthProvider) AddUser(user *domain.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	// Direct access to embedded Casdoor user
	success, err := c.client.AddUser(&user.User)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to add user", err)
	}

	return success, nil
}

// UpdateUser updates existing user
func (c *CasdoorAuthProvider) UpdateUser(user *domain.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	// Direct access to embedded Casdoor user
	success, err := c.client.UpdateUser(&user.User)
	if err != nil {
		return false, newCasdoorErrorWithCause("failed to update user", err)
	}

	return success, nil
}

// DeleteUser deletes user
func (c *CasdoorAuthProvider) DeleteUser(user *domain.User) (bool, error) {
	if c.client == nil {
		return false, newCasdoorError("client not initialized")
	}

	// Direct access to embedded Casdoor user
	success, err := c.client.DeleteUser(&user.User)
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

// Simplified conversion functions

func (c *CasdoorAuthProvider) convertOAuth2TokenToDomainToken(token *oauth2.Token) *domain.Token {
	return &domain.Token{
		AccessToken:  token.AccessToken,
		TokenType:    token.TokenType,
		RefreshToken: token.RefreshToken,
		Expiry:       token.Expiry,
	}
}

func (c *CasdoorAuthProvider) convertCasdoorClaimsToDomainClaims(claims *casdoorsdk.Claims) *domain.Claims {
	domainClaims := &domain.Claims{
		Username:     claims.User.Name,
		Name:         claims.User.DisplayName,
		Email:        claims.User.Email,
		Organization: claims.User.Owner,
		Type:         claims.User.Type,
		Issuer:       claims.Issuer,
		Subject:      claims.Subject,
	}

	// Handle Audience - convert slice to string
	if len(claims.Audience) > 0 {
		domainClaims.Audience = claims.Audience[0]
	}

	// Handle NumericDate fields
	if claims.ExpiresAt != nil {
		domainClaims.ExpiresAt = claims.ExpiresAt.Unix()
	}
	if claims.IssuedAt != nil {
		domainClaims.IssuedAt = claims.IssuedAt.Unix()
	}
	if claims.NotBefore != nil {
		domainClaims.NotBefore = claims.NotBefore.Unix()
	}

	return domainClaims
}

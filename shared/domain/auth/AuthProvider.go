package domain

import (
	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"golang.org/x/oauth2"
)

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Initialize the auth provider with configuration
	Initialize() error

	// Get OAuth authorization URL for user login
	GetAuthURL(state string) string

	// Exchange authorization code for access token
	GetToken(code, state string) (*oauth2.Token, error)

	// Parse and validate JWT token
	ParseToken(token string) (*casdoorsdk.Claims, error)

	// Get user information by username
	GetUser(username string) (*casdoorsdk.User, error)

	// Get all users
	GetUsers() ([]*casdoorsdk.User, error)

	// Create a new user
	AddUser(user *casdoorsdk.User) (bool, error)

	// Update existing user
	UpdateUser(user *casdoorsdk.User) (bool, error)

	// Delete user
	DeleteUser(user *casdoorsdk.User) (bool, error)

	// Verify if provider is healthy
	HealthCheck() error
}

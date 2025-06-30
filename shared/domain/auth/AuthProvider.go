package domain

import (
	"github.com/r0x16/Raidark/shared/domain/model/auth"
)

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Initialize the auth provider with configuration
	Initialize() error

	// Get OAuth authorization URL for user login
	GetAuthURL(state string) string

	// Exchange authorization code for access token
	GetToken(code, state string) (*auth.Token, error)

	// Refresh OAuth token using refresh token
	RefreshToken(refreshToken string) (*auth.Token, error)

	// Parse and validate JWT token
	ParseToken(token string) (*auth.Claims, error)

	// Get user information by username
	GetUser(username string) (*auth.User, error)

	// Get all users
	GetUsers() ([]*auth.User, error)

	// Create a new user
	AddUser(user *auth.User) (bool, error)

	// Update existing user
	UpdateUser(user *auth.User) (bool, error)

	// Delete user
	DeleteUser(user *auth.User) (bool, error)

	// Verify if provider is healthy
	HealthCheck() error
}

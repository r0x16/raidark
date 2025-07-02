package domain

// AuthProvider defines the interface for authentication providers
type AuthProvider interface {
	// Initialize the auth provider with configuration
	Initialize() error

	// Get OAuth authorization URL for user login
	GetAuthURL(state string) string

	// Exchange authorization code for access token
	GetToken(code, state string) (*Token, error)

	// Refresh OAuth token using refresh token
	RefreshToken(refreshToken string) (*Token, error)

	// Parse and validate JWT token
	ParseToken(token string) (*Claims, error)

	// Get user information by username
	GetUser(username string) (*User, error)

	// Get all users
	GetUsers() ([]*User, error)

	// Create a new user
	AddUser(user *User) (bool, error)

	// Update existing user
	UpdateUser(user *User) (bool, error)

	// Delete user
	DeleteUser(user *User) (bool, error)

	// Verify if provider is healthy
	HealthCheck() error
}

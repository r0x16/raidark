package driver

import (
	"errors"
	"fmt"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/r0x16/Raidark/shared/auth/domain"
)

// ArrayAuthProvider implements the AuthProvider interface using an in-memory array
type ArrayAuthProvider struct {
	users []*domain.User
}

// Verify interface implementation
var _ domain.AuthProvider = &ArrayAuthProvider{}

// NewArrayAuthProvider creates a new ArrayAuthProvider instance
func NewArrayAuthProvider() *ArrayAuthProvider {
	return &ArrayAuthProvider{
		users: make([]*domain.User, 0),
	}
}

// Initialize the auth provider with configuration
func (a *ArrayAuthProvider) Initialize() error {
	// Create some test users
	a.createTestUsers()
	return nil
}

// createTestUsers creates some predefined test users
func (a *ArrayAuthProvider) createTestUsers() {
	testUsers := []*domain.User{
		{
			User: casdoorsdk.User{
				Owner:          "test-org",
				Name:           "admin",
				CreatedTime:    time.Now().Format(time.RFC3339),
				UpdatedTime:    time.Now().Format(time.RFC3339),
				Id:             "admin-id",
				Type:           "normal-user",
				Password:       "admin123",
				DisplayName:    "Administrator",
				FirstName:      "Admin",
				LastName:       "User",
				Avatar:         "",
				Email:          "admin@test.com",
				Phone:          "",
				CountryCode:    "US",
				Region:         "US",
				Location:       "Test Location",
				IsAdmin:        true,
				IsForbidden:    false,
				IsDeleted:      false,
				EmailVerified:  true,
				CreatedIp:      "127.0.0.1",
				LastSigninTime: time.Now().Format(time.RFC3339),
				LastSigninIp:   "127.0.0.1",
			},
		},
		{
			User: casdoorsdk.User{
				Owner:          "test-org",
				Name:           "user1",
				CreatedTime:    time.Now().Format(time.RFC3339),
				UpdatedTime:    time.Now().Format(time.RFC3339),
				Id:             "user1-id",
				Type:           "normal-user",
				Password:       "user123",
				DisplayName:    "Test User 1",
				FirstName:      "Test",
				LastName:       "User",
				Avatar:         "",
				Email:          "user1@test.com",
				Phone:          "",
				CountryCode:    "US",
				Region:         "US",
				Location:       "Test Location",
				IsAdmin:        false,
				IsForbidden:    false,
				IsDeleted:      false,
				EmailVerified:  true,
				CreatedIp:      "127.0.0.1",
				LastSigninTime: time.Now().Format(time.RFC3339),
				LastSigninIp:   "127.0.0.1",
			},
		},
		{
			User: casdoorsdk.User{
				Owner:          "test-org",
				Name:           "user2",
				CreatedTime:    time.Now().Format(time.RFC3339),
				UpdatedTime:    time.Now().Format(time.RFC3339),
				Id:             "user2-id",
				Type:           "normal-user",
				Password:       "user123",
				DisplayName:    "Test User 2",
				FirstName:      "Test",
				LastName:       "User 2",
				Avatar:         "",
				Email:          "user2@test.com",
				Phone:          "",
				CountryCode:    "US",
				Region:         "US",
				Location:       "Test Location",
				IsAdmin:        false,
				IsForbidden:    false,
				IsDeleted:      false,
				EmailVerified:  true,
				CreatedIp:      "127.0.0.1",
				LastSigninTime: time.Now().Format(time.RFC3339),
				LastSigninIp:   "127.0.0.1",
			},
		},
	}

	a.users = testUsers
}

// GetAuthURL gets OAuth authorization URL for user login
func (a *ArrayAuthProvider) GetAuthURL(state string) string {
	// For testing purposes, return a mock URL
	return fmt.Sprintf("http://localhost:8080/mock-auth?state=%s", state)
}

// GetToken exchanges authorization code for access token
func (a *ArrayAuthProvider) GetToken(code, state string) (*domain.Token, error) {
	// For testing purposes, return a mock token
	return &domain.Token{
		AccessToken:  "mock-access-token-" + code,
		TokenType:    "Bearer",
		RefreshToken: "mock-refresh-token-" + code,
		Expiry:       time.Now().Add(1 * time.Hour),
	}, nil
}

// RefreshToken refreshes OAuth token using refresh token
func (a *ArrayAuthProvider) RefreshToken(refreshToken string) (*domain.Token, error) {
	// For testing purposes, return a new mock token
	return &domain.Token{
		AccessToken:  "mock-refreshed-access-token",
		TokenType:    "Bearer",
		RefreshToken: refreshToken,
		Expiry:       time.Now().Add(1 * time.Hour),
	}, nil
}

// ParseToken parses and validates JWT token
func (a *ArrayAuthProvider) ParseToken(token string) (*domain.Claims, error) {
	// For testing purposes, return mock claims for admin user
	return &domain.Claims{
		Username:     "admin",
		Name:         "Administrator",
		Email:        "admin@test.com",
		Organization: "test-org",
		Type:         "normal-user",
		Issuer:       "array-auth-provider",
		Subject:      "admin-id",
		Audience:     "test-audience",
		ExpiresAt:    time.Now().Add(1 * time.Hour).Unix(),
		IssuedAt:     time.Now().Unix(),
		NotBefore:    time.Now().Unix(),
	}, nil
}

// GetUser gets user information by username
func (a *ArrayAuthProvider) GetUser(username string) (*domain.User, error) {
	for _, user := range a.users {
		if user.Name == username {
			return user, nil
		}
	}
	return nil, errors.New("user not found")
}

// GetUsers gets all users
func (a *ArrayAuthProvider) GetUsers() ([]*domain.User, error) {
	return a.users, nil
}

// AddUser creates a new user
func (a *ArrayAuthProvider) AddUser(user *domain.User) (bool, error) {
	// Check if user already exists
	for _, existingUser := range a.users {
		if existingUser.Name == user.Name {
			return false, errors.New("user already exists")
		}
	}

	// Set default values if not provided
	if user.Id == "" {
		user.Id = fmt.Sprintf("%s-id", user.Name)
	}
	if user.CreatedTime == "" {
		user.CreatedTime = time.Now().Format(time.RFC3339)
	}
	if user.UpdatedTime == "" {
		user.UpdatedTime = time.Now().Format(time.RFC3339)
	}
	if user.Owner == "" {
		user.Owner = "test-org"
	}
	if user.Type == "" {
		user.Type = "normal-user"
	}

	a.users = append(a.users, user)
	return true, nil
}

// UpdateUser updates existing user
func (a *ArrayAuthProvider) UpdateUser(user *domain.User) (bool, error) {
	for i, existingUser := range a.users {
		if existingUser.Name == user.Name {
			user.UpdatedTime = time.Now().Format(time.RFC3339)
			a.users[i] = user
			return true, nil
		}
	}
	return false, errors.New("user not found")
}

// DeleteUser deletes user
func (a *ArrayAuthProvider) DeleteUser(user *domain.User) (bool, error) {
	for i, existingUser := range a.users {
		if existingUser.Name == user.Name {
			a.users = append(a.users[:i], a.users[i+1:]...)
			return true, nil
		}
	}
	return false, errors.New("user not found")
}

// HealthCheck verifies if provider is healthy
func (a *ArrayAuthProvider) HealthCheck() error {
	if len(a.users) == 0 {
		return errors.New("no users available")
	}
	return nil
}

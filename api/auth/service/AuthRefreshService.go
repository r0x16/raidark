package service

import (
	"fmt"

	"github.com/r0x16/Raidark/api/auth/domain/model"
	"github.com/r0x16/Raidark/api/auth/domain/repositories"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"github.com/r0x16/Raidark/shared/domain/model/auth"
)

// AuthRefreshService handles token refresh functionality
type AuthRefreshService struct {
	*AuthService
}

// NewAuthRefreshService creates a new refresh service
func NewAuthRefreshService(sessionRepo repositories.SessionRepository, authProvider domauth.AuthProvider) *AuthRefreshService {
	return &AuthRefreshService{
		AuthService: NewAuthService(sessionRepo, authProvider),
	}
}

// RefreshTokens refreshes access token using refresh token from session
func (s *AuthRefreshService) RefreshTokens(sessionID, userAgent, ipAddress string) (*model.AuthSession, *auth.Token, error) {
	// Find session by ID using repository
	session, err := s.GetSessionRepo().FindBySessionID(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find session: %w", err)
	}

	// Check if session is expired
	if session.IsRefreshExpired() {
		// Clean up expired session using repository
		s.GetSessionRepo().Delete(session)
		return nil, nil, fmt.Errorf("refresh token expired")
	}

	// Use the auth provider to refresh the token
	newToken, err := s.GetAuthProvider().RefreshToken(session.RefreshToken)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update session with new token information
	session.AccessToken = newToken.AccessToken
	if newToken.RefreshToken != "" {
		session.RefreshToken = newToken.RefreshToken
	}
	session.ExpiresAt = newToken.Expiry
	session.UserAgent = userAgent
	session.IPAddress = ipAddress

	// Save updated session to database
	if err := s.GetSessionRepo().Update(session); err != nil {
		return nil, nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, newToken, nil
}

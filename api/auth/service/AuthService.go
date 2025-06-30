package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/casdoor/casdoor-go-sdk/casdoorsdk"
	"github.com/r0x16/Raidark/api/auth/domain/model"
	"github.com/r0x16/Raidark/api/auth/domain/repositories"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"golang.org/x/oauth2"
	"gorm.io/gorm"
)

// AuthService handles authentication business logic
type AuthService struct {
	sessionRepo  repositories.SessionRepository
	authProvider domauth.AuthProvider
}

// NewAuthService creates a new authentication service
func NewAuthService(sessionRepo repositories.SessionRepository, authProvider domauth.AuthProvider) *AuthService {
	return &AuthService{
		sessionRepo:  sessionRepo,
		authProvider: authProvider,
	}
}

// ExchangeCodeForTokens exchanges authorization code for tokens and creates session
func (s *AuthService) ExchangeCodeForTokens(code, state, userAgent, ipAddress string) (*model.AuthSession, *oauth2.Token, *casdoorsdk.Claims, error) {
	// Exchange code for token using Casdoor
	token, err := s.authProvider.GetToken(code, state)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Parse JWT token to get user claims
	claims, err := s.authProvider.ParseToken(token.AccessToken)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to parse JWT token: %w", err)
	}

	// Generate unique session ID
	sessionID, err := s.generateSessionID()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	// Create session record
	session := &model.AuthSession{
		SessionID:     sessionID,
		UserID:        claims.User.Id,
		Username:      claims.User.Name,
		RefreshToken:  token.RefreshToken,
		AccessToken:   token.AccessToken,
		ExpiresAt:     token.Expiry,
		RefreshExpiry: time.Now().Add(30 * 24 * time.Hour), // 30 days for refresh token
		UserAgent:     userAgent,
		IPAddress:     ipAddress,
	}

	// Save session to database using repository
	if err := s.sessionRepo.Create(session); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, token, claims, nil
}

// RefreshTokens refreshes access token using refresh token from session
func (s *AuthService) RefreshTokens(sessionID, userAgent, ipAddress string) (*model.AuthSession, *oauth2.Token, error) {
	// Find session by ID using repository
	session, err := s.sessionRepo.FindBySessionID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("session not found")
		}
		return nil, nil, fmt.Errorf("failed to find session: %w", err)
	}

	// Check if session is expired
	if session.IsRefreshExpired() {
		// Clean up expired session using repository
		s.sessionRepo.Delete(session)
		return nil, nil, fmt.Errorf("refresh token expired")
	}

	// Note: Casdoor SDK doesn't have direct refresh token method
	// We need to handle this manually by making HTTP request to Casdoor
	// For now, we'll return an error indicating this needs to be implemented
	return nil, nil, fmt.Errorf("refresh token functionality not yet implemented - requires custom HTTP client for Casdoor")
}

// InvalidateSession removes session from database (logout)
func (s *AuthService) InvalidateSession(sessionID string) error {
	// Find session first to ensure it exists
	session, err := s.sessionRepo.FindBySessionID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to find session: %w", err)
	}

	// Delete session from database using repository
	if err := s.sessionRepo.Delete(session); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

// GetSessionByID retrieves session by session ID
func (s *AuthService) GetSessionByID(sessionID string) (*model.AuthSession, error) {
	session, err := s.sessionRepo.FindBySessionID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to find session: %w", err)
	}

	return session, nil
}

// GetUserSessions retrieves all sessions for a user
func (s *AuthService) GetUserSessions(userID string) ([]*model.AuthSession, error) {
	sessions, err := s.sessionRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user sessions: %w", err)
	}

	return sessions, nil
}

// InvalidateAllUserSessions removes all sessions for a user (security operation)
func (s *AuthService) InvalidateAllUserSessions(userID string) error {
	if err := s.sessionRepo.DeleteAllByUserID(userID); err != nil {
		return fmt.Errorf("failed to delete all user sessions: %w", err)
	}

	return nil
}

// generateSessionID creates a unique session identifier
func (s *AuthService) generateSessionID() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// CleanExpiredSessions removes expired sessions from database
func (s *AuthService) CleanExpiredSessions() error {
	return s.sessionRepo.DeleteExpiredSessions()
}

package service

import (
	"fmt"
	"time"

	"github.com/r0x16/Raidark/api/auth/domain/model"
	"github.com/r0x16/Raidark/api/auth/domain/repositories"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"github.com/r0x16/Raidark/shared/domain/model/auth"
)

// AuthExchangeService handles code exchange for tokens
type AuthExchangeService struct {
	*AuthService
}

// NewAuthExchangeService creates a new exchange service
func NewAuthExchangeService(sessionRepo repositories.SessionRepository, authProvider domauth.AuthProvider) *AuthExchangeService {
	return &AuthExchangeService{
		AuthService: NewAuthService(sessionRepo, authProvider),
	}
}

// ExchangeCodeForTokens exchanges authorization code for tokens and creates session
func (s *AuthExchangeService) ExchangeCodeForTokens(code, state, userAgent, ipAddress string) (*model.AuthSession, *auth.Token, *auth.Claims, error) {
	// Exchange code for token using Casdoor
	token, err := s.GetAuthProvider().GetToken(code, state)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to exchange code for token: %w", err)
	}

	// Parse JWT token to get user claims
	claims, err := s.GetAuthProvider().ParseToken(token.AccessToken)
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
		UserID:        claims.Subject,
		Username:      claims.Username,
		RefreshToken:  token.RefreshToken,
		AccessToken:   token.AccessToken,
		ExpiresAt:     token.Expiry,
		RefreshExpiry: time.Now().Add(30 * 24 * time.Hour), // 30 days for refresh token
		UserAgent:     userAgent,
		IPAddress:     ipAddress,
	}

	// Save session to database using repository
	if err := s.GetSessionRepo().Create(session); err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, token, claims, nil
}

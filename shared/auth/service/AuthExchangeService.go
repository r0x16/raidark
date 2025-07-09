package service

import (
	"fmt"
	"time"

	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/domain/event"
	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/domain/repositories"
	domevents "github.com/r0x16/Raidark/shared/events/domain"
)

// AuthExchangeService handles code exchange for tokens
type AuthExchangeService struct {
	*AuthService
	events domevents.DomainEventsProvider
}

// NewAuthExchangeService creates a new exchange service
func NewAuthExchangeService(
	sessionRepo repositories.SessionRepository,
	authProvider domain.AuthProvider,
	events domevents.DomainEventsProvider,
) *AuthExchangeService {
	authService := NewAuthService(sessionRepo, authProvider)
	return &AuthExchangeService{
		AuthService: authService,
		events:      events,
	}
}

// ExchangeCodeForTokens exchanges authorization code for tokens and creates session
func (s *AuthExchangeService) ExchangeCodeForTokens(code, state, userAgent, ipAddress string) (*model.AuthSession, *domain.Token, *domain.Claims, error) {
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

	// Publish event
	if s.events != nil {
		s.events.Publish(&event.SessionWasCreated{
			Session: session,
		})
	}
	return session, token, claims, nil
}

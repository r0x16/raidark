package service

import (
	"fmt"
	"time"

	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/domain/event"
	"github.com/r0x16/Raidark/shared/auth/domain/repositories"
	domevents "github.com/r0x16/Raidark/shared/events/domain"
)

// AuthLogoutService handles logout functionality
type AuthLogoutService struct {
	*AuthService
	events domevents.DomainEventsProvider
}

// NewAuthLogoutService creates a new logout service
func NewAuthLogoutService(
	sessionRepo repositories.SessionRepository,
	authProvider domain.AuthProvider,
	events domevents.DomainEventsProvider,
) *AuthLogoutService {
	authService := NewAuthService(sessionRepo, authProvider)
	return &AuthLogoutService{
		AuthService: authService,
		events:      events,
	}
}

// InvalidateSession removes session from database (logout)
func (s *AuthLogoutService) InvalidateSession(sessionID string) error {
	// Find session first to ensure it exists
	session, err := s.GetSessionRepo().FindBySessionID(sessionID)
	if err != nil {
		return fmt.Errorf("failed to find session: %w", err)
	}

	// Delete session from database using repository
	if err := s.GetSessionRepo().Delete(session); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	if s.events != nil {
		s.events.Publish(&event.SessionWasDeleted{
			Session:  session,
			LogoutAt: time.Now(),
		})
	}

	return nil
}

package service

import (
	"fmt"

	"github.com/r0x16/Raidark/api/auth/domain/repositories"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"gorm.io/gorm"
)

// AuthLogoutService handles logout functionality
type AuthLogoutService struct {
	*AuthService
}

// NewAuthLogoutService creates a new logout service
func NewAuthLogoutService(sessionRepo repositories.SessionRepository, authProvider domauth.AuthProvider) *AuthLogoutService {
	return &AuthLogoutService{
		AuthService: NewAuthService(sessionRepo, authProvider),
	}
}

// InvalidateSession removes session from database (logout)
func (s *AuthLogoutService) InvalidateSession(sessionID string) error {
	// Find session first to ensure it exists
	session, err := s.GetSessionRepo().FindBySessionID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to find session: %w", err)
	}

	// Delete session from database using repository
	if err := s.GetSessionRepo().Delete(session); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

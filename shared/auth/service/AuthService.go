package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/auth/domain/repositories"
)

// AuthService base service with common authentication functionalities
type AuthService struct {
	sessionRepo  repositories.SessionRepository
	authProvider domain.AuthProvider
}

// NewAuthService creates a new base authentication service
func NewAuthService(sessionRepo repositories.SessionRepository, authProvider domain.AuthProvider) *AuthService {
	return &AuthService{
		sessionRepo:  sessionRepo,
		authProvider: authProvider,
	}
}

// GetSessionByID retrieves session by session ID
func (s *AuthService) GetSessionByID(sessionID string) (*model.AuthSession, error) {
	session, err := s.sessionRepo.FindBySessionID(sessionID)
	if err != nil {
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

// GetSessionRepo returns the session repository for access by specialized services
func (s *AuthService) GetSessionRepo() repositories.SessionRepository {
	return s.sessionRepo
}

// GetAuthProvider returns the auth provider for access by specialized services
func (s *AuthService) GetAuthProvider() domain.AuthProvider {
	return s.authProvider
}

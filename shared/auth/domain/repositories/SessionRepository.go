package repositories

import "github.com/r0x16/Raidark/shared/auth/domain/model"

// SessionRepository defines the interface for session data access operations
type SessionRepository interface {
	// Create a new session record
	Create(session *model.AuthSession) error

	// Find session by session ID
	FindBySessionID(sessionID string) (*model.AuthSession, error)

	// Update existing session
	Update(session *model.AuthSession) error

	// Delete session by session ID
	DeleteBySessionID(sessionID string) error

	// Delete session record
	Delete(session *model.AuthSession) error

	// Find all expired sessions for cleanup
	FindExpiredSessions() ([]*model.AuthSession, error)

	// Delete all expired sessions
	DeleteExpiredSessions() error

	// Find sessions by user ID
	FindByUserID(userID string) ([]*model.AuthSession, error)

	// Delete all sessions for a user (useful for security operations)
	DeleteAllByUserID(userID string) error
}

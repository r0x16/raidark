package repositories

import (
	"time"

	"github.com/r0x16/Raidark/api/auth/domain/model"
	"github.com/r0x16/Raidark/api/auth/domain/repositories"
	"gorm.io/gorm"
)

// GormSessionRepository implements SessionRepository using GORM
type GormSessionRepository struct {
	db *gorm.DB
}

// Verify interface implementation
var _ repositories.SessionRepository = &GormSessionRepository{}

// NewGormSessionRepository creates a new GORM session repository instance
func NewGormSessionRepository(db *gorm.DB) *GormSessionRepository {
	return &GormSessionRepository{
		db: db,
	}
}

// Create implements repositories.SessionRepository
func (r *GormSessionRepository) Create(session *model.AuthSession) error {
	return r.db.Create(session).Error
}

// FindBySessionID implements repositories.SessionRepository
func (r *GormSessionRepository) FindBySessionID(sessionID string) (*model.AuthSession, error) {
	var session model.AuthSession
	err := r.db.Where("session_id = ?", sessionID).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update implements repositories.SessionRepository
func (r *GormSessionRepository) Update(session *model.AuthSession) error {
	return r.db.Save(session).Error
}

// DeleteBySessionID implements repositories.SessionRepository
func (r *GormSessionRepository) DeleteBySessionID(sessionID string) error {
	return r.db.Where("session_id = ?", sessionID).Delete(&model.AuthSession{}).Error
}

// Delete implements repositories.SessionRepository
func (r *GormSessionRepository) Delete(session *model.AuthSession) error {
	return r.db.Delete(session).Error
}

// FindExpiredSessions implements repositories.SessionRepository
func (r *GormSessionRepository) FindExpiredSessions() ([]*model.AuthSession, error) {
	var sessions []*model.AuthSession
	now := time.Now()
	err := r.db.Where("refresh_expiry < ?", now).Find(&sessions).Error
	return sessions, err
}

// DeleteExpiredSessions implements repositories.SessionRepository
func (r *GormSessionRepository) DeleteExpiredSessions() error {
	now := time.Now()
	return r.db.Where("refresh_expiry < ?", now).Delete(&model.AuthSession{}).Error
}

// FindByUserID implements repositories.SessionRepository
func (r *GormSessionRepository) FindByUserID(userID string) ([]*model.AuthSession, error) {
	var sessions []*model.AuthSession
	err := r.db.Where("user_id = ?", userID).Find(&sessions).Error
	return sessions, err
}

// DeleteAllByUserID implements repositories.SessionRepository
func (r *GormSessionRepository) DeleteAllByUserID(userID string) error {
	return r.db.Where("user_id = ?", userID).Delete(&model.AuthSession{}).Error
}

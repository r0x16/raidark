package model

import (
	"time"

	"github.com/r0x16/Raidark/domain/model"
)

// AuthSession represents a user authentication session in the database
type AuthSession struct {
	model.BaseModel
	SessionID     string    `gorm:"uniqueIndex;type:varchar(255);not null" json:"session_id"`
	UserID        string    `gorm:"type:varchar(255);not null" json:"user_id"`
	Username      string    `gorm:"type:varchar(255);not null" json:"username"`
	RefreshToken  string    `gorm:"type:text;not null" json:"refresh_token"`
	AccessToken   string    `gorm:"type:text;not null" json:"access_token"`
	ExpiresAt     time.Time `gorm:"type:timestamp;not null" json:"expires_at"`
	RefreshExpiry time.Time `gorm:"type:timestamp;not null" json:"refresh_expiry"`
	UserAgent     string    `gorm:"type:varchar(500)" json:"user_agent"`
	IPAddress     string    `gorm:"type:varchar(45)" json:"ip_address"`
}

// StoreName returns the datastore name for GORM
func (AuthSession) StoreName() string {
	return "auth_sessions"
}

// IsExpired checks if the session has expired
func (s *AuthSession) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsRefreshExpired checks if the refresh token has expired
func (s *AuthSession) IsRefreshExpired() bool {
	return time.Now().After(s.RefreshExpiry)
}

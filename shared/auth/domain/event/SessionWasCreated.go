package event

import (
	"time"

	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/events/domain"
)

// SessionWasCreated is an event that is published when a session is created
type SessionWasCreated struct {
	Session *model.AuthSession
}

var _ domain.DomainEvent = &SessionWasCreated{}

func (e *SessionWasCreated) Name() string {
	return "auth.session.created"
}

func (e *SessionWasCreated) OccurredAt() time.Time {
	return e.Session.CreatedAt
}

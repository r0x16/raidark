package event

import (
	"time"

	"github.com/r0x16/Raidark/shared/auth/domain/model"
	"github.com/r0x16/Raidark/shared/events/domain"
)

type SessionWasDeleted struct {
	Session  *model.AuthSession
	LogoutAt time.Time
}

var _ domain.DomainEvent = &SessionWasDeleted{}

func (e *SessionWasDeleted) Name() string {
	return "auth.session.deleted"
}

func (e *SessionWasDeleted) OccurredAt() time.Time {
	return e.LogoutAt
}

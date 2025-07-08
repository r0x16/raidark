package domain

import (
	"github.com/r0x16/Raidark/shared/api/domain"
)

type EventClient interface {
	GetId() string
	Setup() *domain.Error
	SendMessage(message *EventMessage) *domain.Error
	Online() *domain.Error
}

package events

import (
	"github.com/r0x16/Raidark/shared/domain/output"
)

type EventClient interface {
	GetId() string
	Setup() *output.Error
	SendMessage(message *EventMessage) *output.Error
	Online() *output.Error
}

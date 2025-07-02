package domain

import "github.com/r0x16/Raidark/shared/domain/output"

type ServerEventProvider interface {
	Subscribe(client EventClient) *output.Error
	Unsubscribe(client EventClient) *output.Error
	Broadcast(message *EventMessage) *output.Error
}

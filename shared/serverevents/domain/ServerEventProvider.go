package domain

import "github.com/r0x16/Raidark/shared/api/domain"

type ServerEventProvider interface {
	Subscribe(client EventClient) *domain.Error
	Unsubscribe(client EventClient) *domain.Error
	Broadcast(message *EventMessage) *domain.Error
}

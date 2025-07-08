package driver

import (
	"net/http"
	"sync"

	domapi "github.com/r0x16/Raidark/shared/api/domain"
	"github.com/r0x16/Raidark/shared/serverevents/domain"
)

type ServerEventEcho struct {
	EventId string
	Clients map[string]EventClientEcho
	m       sync.Mutex
}

var _ domain.ServerEventProvider = &ServerEventEcho{}

// NewServerEventEcho creates a new ServerEventEcho instance.
func NewServerEventEcho(id string) *ServerEventEcho {
	return &ServerEventEcho{
		EventId: id,
		Clients: make(map[string]EventClientEcho),
	}
}

// Subscribe implements events.ServerEventProvider.
func (se *ServerEventEcho) Subscribe(client domain.EventClient) *domapi.Error {
	se.m.Lock()
	defer se.m.Unlock()

	id := client.GetId()

	if _, ok := se.Clients[id]; ok {
		return &domapi.Error{
			Code:    http.StatusInternalServerError,
			Message: "client already subscribed",
		}
	}

	se.Clients[id] = client.(EventClientEcho)
	return nil
}

// Unsubscribe implements events.ServerEventProvider.
func (se *ServerEventEcho) Unsubscribe(client domain.EventClient) *domapi.Error {
	se.m.Lock()
	defer se.m.Unlock()

	id := client.GetId()

	if _, ok := se.Clients[id]; !ok {
		return &domapi.Error{
			Code:    http.StatusInternalServerError,
			Message: "client not subscribed",
		}
	}

	delete(se.Clients, id)
	return nil
}

// Broadcast implements events.ServerEventProvider.
func (se *ServerEventEcho) Broadcast(message *domain.EventMessage) *domapi.Error {
	se.m.Lock()
	defer se.m.Unlock()

	for _, client := range se.Clients {
		err := client.SendMessage(message)
		if err != nil {
			return err
		}
	}

	return nil
}

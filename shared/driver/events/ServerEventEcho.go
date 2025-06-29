package events

import (
	"net/http"
	"sync"

	"github.com/r0x16/Raidark/shared/domain/events"
	"github.com/r0x16/Raidark/shared/domain/output"
)

type ServerEventEcho struct {
	EventId string
	Clients map[string]EventClientEcho
	m       sync.Mutex
}

var _ events.ServerEventProvider = &ServerEventEcho{}

// NewServerEventEcho creates a new ServerEventEcho instance.
func NewServerEventEcho(id string) *ServerEventEcho {
	return &ServerEventEcho{
		EventId: id,
		Clients: make(map[string]EventClientEcho),
	}
}

// Subscribe implements events.ServerEventProvider.
func (se *ServerEventEcho) Subscribe(client events.EventClient) *output.Error {
	se.m.Lock()
	defer se.m.Unlock()

	id := client.GetId()

	if _, ok := se.Clients[id]; ok {
		return &output.Error{
			Code:    http.StatusInternalServerError,
			Message: "client already subscribed",
		}
	}

	se.Clients[id] = client.(EventClientEcho)
	return nil
}

// Unsubscribe implements events.ServerEventProvider.
func (se *ServerEventEcho) Unsubscribe(client events.EventClient) *output.Error {
	se.m.Lock()
	defer se.m.Unlock()

	id := client.GetId()

	if _, ok := se.Clients[id]; !ok {
		return &output.Error{
			Code:    http.StatusInternalServerError,
			Message: "client not subscribed",
		}
	}

	delete(se.Clients, id)
	return nil
}

// Broadcast implements events.ServerEventProvider.
func (se *ServerEventEcho) Broadcast(message *events.EventMessage) *output.Error {
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

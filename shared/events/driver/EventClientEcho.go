package events

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/domain/output"
	"github.com/r0x16/Raidark/shared/events/domain"
)

type EventClientEcho struct {
	Id           string
	eventChannel chan *domain.EventMessage
	context      echo.Context
}

var _ domain.EventClient = &EventClientEcho{}

// NewEventClientEcho creates a new EventClientEcho instance.
func NewEventClientEcho(id string, c echo.Context) EventClientEcho {
	return EventClientEcho{
		Id:           id,
		eventChannel: make(chan *domain.EventMessage, 1),
		context:      c,
	}
}

func (c EventClientEcho) Setup() *output.Error {
	c.context.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.context.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type")
	c.context.Response().Header().Set("Content-Type", "text/event-stream")
	c.context.Response().Header().Set("Cache-Control", "no-cache")
	c.context.Response().Header().Set("Connection", "keep-alive")
	return nil
}

// GetId implements domain.EventClient.
func (c EventClientEcho) GetId() string {
	return c.Id
}

// SendMessage implements domain.EventClient.
func (c EventClientEcho) SendMessage(message *domain.EventMessage) *output.Error {
	c.eventChannel <- message
	return nil
}

// WaitForMessage implements domain.EventClient.
func (c EventClientEcho) Online() *output.Error {
	c.ping()
	for {
		select {
		case message := <-c.eventChannel:
			err := c.handleEvent(message)
			if err != nil {
				return err
			}
		case <-c.context.Request().Context().Done():
			c.Close()
			return nil
		}
	}
}

func (c EventClientEcho) Close() {
	c.handleEvent(&domain.EventMessage{
		Event: "ready",
		Data:  "close",
	})
	close(c.eventChannel)
}

func (c EventClientEcho) handleEvent(message *domain.EventMessage) *output.Error {
	data, err := json.Marshal(message.Data)
	if err != nil {
		return &output.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error processing event data",
			Data:    err,
		}
	}

	return c.transportEvent(message.Event, string(data))

}

func (c EventClientEcho) transportEvent(event string, data string) *output.Error {
	const format = "event:%s\ndata:%s\n\n"
	_, err := c.context.Response().Writer.Write([]byte(fmt.Sprintf(format, event, data)))
	if err != nil {
		return &output.Error{
			Code:    http.StatusInternalServerError,
			Message: "Error sending event",
			Data:    err,
		}
	}

	c.context.Response().Flush()
	return nil
}

func (c EventClientEcho) ping() *output.Error {
	return c.handleEvent(&domain.EventMessage{
		Event: "ping",
		Data:  "pong",
	})
}

package driver

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	domain "github.com/r0x16/Raidark/shared/serverevents/domain"
)

func newClient(t *testing.T) EventClientEcho {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	client := NewEventClientEcho("id", c)
	return client
}

func TestServerEventEchoSubscribeUnsubscribe(t *testing.T) {
	se := NewServerEventEcho("ev")
	client := newClient(t)
	if err := se.Subscribe(client); err != nil {
		t.Fatalf("subscribe err %v", err)
	}
	if err := se.Unsubscribe(client); err != nil {
		t.Fatalf("unsubscribe err %v", err)
	}
}

func TestServerEventEchoBroadcast(t *testing.T) {
	se := NewServerEventEcho("ev")
	client := newClient(t)
	se.Subscribe(client)
	msg := &domain.EventMessage{Event: "ready", Data: "x"}
	if err := se.Broadcast(msg); err != nil {
		t.Fatalf("broadcast err %v", err)
	}
}

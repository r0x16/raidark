package driver

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSubscribeUnsubscribe(t *testing.T) {
	se := NewServerEventEcho("e")
	e := echo.New()
	c := e.NewContext(nil, httptest.NewRecorder())
	client := NewEventClientEcho("1", c)
	if err := se.Subscribe(client); err != nil {
		t.Fatalf("subscribe err: %v", err)
	}
	if err := se.Unsubscribe(client); err != nil {
		t.Fatalf("unsubscribe err: %v", err)
	}
}

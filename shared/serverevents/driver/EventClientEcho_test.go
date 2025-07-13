package driver

import (
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	domain "github.com/r0x16/Raidark/shared/serverevents/domain"
)

func TestEventClientEchoSetupAndSend(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	client := NewEventClientEcho("id", c)
	if err := client.Setup(); err != nil {
		t.Fatalf("setup err %v", err)
	}
	if err := client.SendMessage(&domain.EventMessage{Event: "ready", Data: "test"}); err != nil {
		t.Fatalf("send err %v", err)
	}
}

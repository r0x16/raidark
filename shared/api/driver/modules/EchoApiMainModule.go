package modules

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/domain"
)

type EchoApiMainModule struct {
	*EchoModule
}

var _ domain.ApiModule = &EchoApiMainModule{}

// Name implements domain.ApiModule.
func (e *EchoApiMainModule) Name() string {
	return "ApiMain"
}

// Setup implements domain.ApiModule.
func (e *EchoApiMainModule) Setup() error {

	e.Group.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	return nil
}

func (e *EchoApiMainModule) GetModel() []any {
	return []any{}
}

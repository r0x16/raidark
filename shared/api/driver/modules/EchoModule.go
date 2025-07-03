package modules

import (
	"github.com/labstack/echo/v4"
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type EchoModule struct {
	Api   domapi.ApiProvider
	Group *echo.Group
	Hub   *domprovider.ProviderHub
}

type ActionCallback func(echo.Context, *domprovider.ProviderHub) error

func (e *EchoModule) ActionInjection(callback ActionCallback) echo.HandlerFunc {
	if e.Hub == nil {
		panic("Hub is not set in EchoModule")
	}

	return func(c echo.Context) error {
		return callback(c, e.Hub)
	}
}

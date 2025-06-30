package modules

import (
	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/api/drivers"
)

type EchoModule struct {
	Api   *drivers.EchoApiProvider
	Group *echo.Group
}

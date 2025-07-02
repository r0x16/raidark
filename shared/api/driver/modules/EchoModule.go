package modules

import (
	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/driver"
)

type EchoModule struct {
	Api   *driver.EchoApiProvider
	Group *echo.Group
}

package drivers

import (
	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/driver/db"
	"github.com/r0x16/Raidark/shared/driver/events"
)

type ApplicationBundle struct {
	Log      *events.StdOutLogManager
	Database *db.GormMysqlDatabaseProvider
}

type ActionCallback func(echo.Context, *ApplicationBundle) error

func (bundle *ApplicationBundle) ActionInjection(callback ActionCallback) echo.HandlerFunc {
	return func(c echo.Context) error {
		return callback(c, bundle)
	}
}

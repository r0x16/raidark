package drivers

import (
	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/domain"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"github.com/r0x16/Raidark/shared/domain/logger"
	"github.com/r0x16/Raidark/shared/driver/env"
)

/*
 * Represents the application bundle
 * This bundle contains the log manager, database provider, auth provider, and environment provider
 */
type ApplicationBundle struct {
	Log      logger.LogProvider
	Database domain.DatabaseProvider
	Auth     domauth.AuthProvider
	Env      *env.EnvProvider
}

/*
 * Represents the action callback
 * This callback is used to inject the application bundle into the action
 */
type ActionCallback func(echo.Context, *ApplicationBundle) error

/*
 * Injects the application bundle into the action
 * This method is used to inject the application bundle into the action
 */
func (bundle *ApplicationBundle) ActionInjection(callback ActionCallback) echo.HandlerFunc {
	return func(c echo.Context) error {
		return callback(c, bundle)
	}
}

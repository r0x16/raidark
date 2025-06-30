package modules

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/api/domain"
	"github.com/r0x16/Raidark/api/drivers"
)

type EchoMainModule struct {
	Api *drivers.EchoApiProvider
}

var _ domain.ApiModule = &EchoMainModule{}

// Name implements domain.ApiModule.
func (e *EchoMainModule) Name() string {
	return "Main"
}

// Setup implements domain.ApiModule.
func (e *EchoMainModule) Setup() error {
	e.Api.Server.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// CSRF token endpoint for frontend applications
	e.Api.Server.GET("/csrf-token", func(c echo.Context) error {
		// The CSRF token is automatically available in the context when CSRF middleware is enabled
		token := c.Get("csrf")
		if token == nil {
			// CSRF middleware is disabled
			return c.JSON(http.StatusOK, map[string]string{
				"message": "CSRF protection is disabled",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"csrf_token": token.(string),
		})
	})

	return nil
}

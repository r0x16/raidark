package modules

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/domain"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type EchoMainModule struct {
	*EchoModule
}

var _ domain.ApiModule = &EchoMainModule{}

// Name implements domain.ApiModule.
func (e *EchoMainModule) Name() string {
	return "Main"
}

// Setup implements domain.ApiModule.
func (e *EchoMainModule) Setup() error {

	env := domprovider.Get[domenv.EnvProvider](e.Hub)

	e.Group.GET("/health", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	// CSRF token endpoint for frontend applications
	e.Group.GET("/csrf-token", func(c echo.Context) error {
		// Check if CSRF is enabled
		csrfEnabled := env.GetBool("CSRF_ENABLED", false)
		if !csrfEnabled {
			return c.JSON(http.StatusNotFound, map[string]string{
				"error": "CSRF protection is disabled",
			})
		}

		// The CSRF token is automatically available in the context when CSRF middleware is enabled
		token := c.Get("csrf")
		if token == nil {
			// CSRF middleware is disabled (shouldn't happen due to check above, but for safety)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "CSRF token not available",
			})
		}

		return c.JSON(http.StatusOK, map[string]string{
			"csrf_token": token.(string),
		})
	})

	return nil
}

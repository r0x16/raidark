package modules

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/domain"
	"github.com/r0x16/Raidark/shared/api/rest"
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

	// /csrf-token is only registered when CSRF_ENABLED=true. When disabled the route simply
	// does not exist, so callers receive Echo's 404 (route not found) rather than a custom
	// 404 from the handler. This matches the principle: disabled features leave no surface.
	if env.GetBool("CSRF_ENABLED", false) {
		e.Group.GET("/csrf-token", func(c echo.Context) error {
			token := c.Get("csrf")
			if token == nil {
				return rest.RenderError(c, http.StatusInternalServerError, &rest.RESTError{
					Code:    "csrf.unavailable",
					Message: "CSRF token not available.",
				})
			}
			return c.JSON(http.StatusOK, map[string]string{
				"csrf_token": token.(string),
			})
		})
	}

	return nil
}

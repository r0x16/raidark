package rest

import (
	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/ids"
)

const (
	// correlationIDKey is the echo.Context key used to store and retrieve the ID.
	correlationIDKey = "correlation_id"

	// correlationIDHeader is the HTTP header name for the correlation ID.
	correlationIDHeader = "X-Correlation-ID"
)

// CorrelationID returns an Echo middleware that propagates a request-scoped
// correlation ID across service boundaries. The middleware reads X-Correlation-ID
// from the incoming request; if absent or empty, it generates a new UUIDv7.
// The resolved ID is stored in echo.Context and echoed back in the response header
// so callers can use it to correlate distributed traces and log entries.
func CorrelationID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			id := c.Request().Header.Get(correlationIDHeader)
			if id == "" {
				generated, err := ids.NewV7()
				if err != nil {
					// NewV7 failure is exceedingly rare (OS entropy exhausted).
					// Use a zero UUID as a last-resort fallback rather than rejecting the request.
					generated = "00000000-0000-7000-8000-000000000000"
				}
				id = generated
			}
			c.Set(correlationIDKey, id)
			c.Response().Header().Set(correlationIDHeader, id)
			return next(c)
		}
	}
}

// GetCorrelationID retrieves the correlation ID injected by the CorrelationID middleware.
// Returns an empty string if the middleware was not installed for the current route.
func GetCorrelationID(c echo.Context) string {
	v, _ := c.Get(correlationIDKey).(string)
	return v
}

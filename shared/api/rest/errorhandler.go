package rest

import (
	"github.com/labstack/echo/v4"
)

// EchoErrorHandler is a custom Echo HTTP error handler that converts any unhandled
// error returned by a handler into the standard REST envelope.
//
// In Raidark, errors are expressed using our own sentinels (rest.ErrNotFound, etc.)
// or by calling rest.RenderError directly. Echo's native echo.HTTPError is not used.
//
// Flow:
//   - Handler calls rest.RenderError → response already committed → this handler skips.
//   - Handler returns a sentinel error → MapError resolves status + code → rendered here.
//   - Handler returns any other error → 500 / internal.unexpected → rendered here.
//
// Register with: e.HTTPErrorHandler = rest.EchoErrorHandler
func EchoErrorHandler(err error, c echo.Context) {
	if c.Response().Committed {
		return
	}
	status, restErr := MapError(err)
	_ = RenderError(c, status, restErr)
}

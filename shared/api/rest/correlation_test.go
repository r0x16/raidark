// Package rest_test verifies Raidark's public REST correlation ID middleware
// contract from the point of view of services and HTTP clients.
package rest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/rest"
	"github.com/r0x16/Raidark/shared/ids"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCorrelationID_headerPresentIsRespected verifies callers can provide their
// own correlation ID and see the same value in handlers and response headers.
func TestCorrelationID_headerPresentIsRespected(t *testing.T) {
	const inputCorrelationID = "client-provided-correlation"

	recorder, captured := runCorrelationMiddleware(t, map[string]string{
		"X-Correlation-ID": inputCorrelationID,
	})

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, inputCorrelationID, captured)
	assert.Equal(t, inputCorrelationID, recorder.Header().Get("X-Correlation-ID"))
}

// TestCorrelationID_headerAbsentGeneratesUUIDv7 verifies the middleware creates
// a valid UUIDv7 correlation ID when the caller omits the header.
func TestCorrelationID_headerAbsentGeneratesUUIDv7(t *testing.T) {
	recorder, captured := runCorrelationMiddleware(t, nil)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.True(t, ids.IsValidV7(captured))
	assert.Equal(t, captured, recorder.Header().Get("X-Correlation-ID"))
}

func runCorrelationMiddleware(t *testing.T, headers map[string]string) (*httptest.ResponseRecorder, string) {
	t.Helper()

	e := echo.New()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	for key, value := range headers {
		request.Header.Set(key, value)
	}
	recorder := httptest.NewRecorder()
	context := e.NewContext(request, recorder)

	var captured string
	handler := rest.CorrelationID()(func(c echo.Context) error {
		captured = rest.GetCorrelationID(c)
		return c.NoContent(http.StatusNoContent)
	})
	require.NoError(t, handler(context))

	return recorder, captured
}

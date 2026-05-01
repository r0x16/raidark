// Package driver_test verifies API provider wiring that affects HTTP behavior
// visible to services built on Raidark.
package driver_test

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
	apidriver "github.com/r0x16/Raidark/shared/api/driver"
	"github.com/r0x16/Raidark/shared/api/rest"
	envdomain "github.com/r0x16/Raidark/shared/env/domain"
	logdomain "github.com/r0x16/Raidark/shared/logger/domain"
	providerdomain "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEchoApiProviderSetup_installsRESTConventions covers the integration added
// by RDK-002: all Echo requests receive correlation IDs and unhandled Raidark
// sentinels are rendered through the REST envelope handler.
func TestEchoApiProviderSetup_installsRESTConventions(t *testing.T) {
	provider := newTestEchoAPIProvider()

	require.NoError(t, provider.Setup())
	provider.Server.GET("/private", func(c echo.Context) error {
		return rest.ErrForbidden
	})

	request := httptest.NewRequest(http.MethodGet, "/private", nil)
	request.Header.Set("X-Correlation-ID", snapshotTraceID)
	recorder := httptest.NewRecorder()

	provider.Server.ServeHTTP(recorder, request)

	assert.Equal(t, handlerPointer(rest.EchoErrorHandler), handlerPointer(provider.Server.HTTPErrorHandler))
	assert.Equal(t, snapshotTraceID, recorder.Header().Get("X-Correlation-ID"))
	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "common.forbidden",
			"message": "You do not have permission to perform this action.",
			"trace_id": "trace-rdk-002"
		}
	}`, recorder.Body.String())
}

const snapshotTraceID = "trace-rdk-002"

type testEnvProvider struct{}

func (testEnvProvider) GetString(_ string, defaultValue string) string { return defaultValue }
func (testEnvProvider) GetBool(_ string, defaultValue bool) bool       { return defaultValue }
func (testEnvProvider) GetInt(_ string, defaultValue int) int          { return defaultValue }
func (testEnvProvider) GetFloat(_ string, defaultValue float64) float64 {
	return defaultValue
}
func (testEnvProvider) GetSlice(_ string, defaultValue []string) []string {
	return defaultValue
}
func (testEnvProvider) GetSliceWithSeparator(_ string, _ string, defaultValue []string) []string {
	return defaultValue
}
func (testEnvProvider) IsSet(_ string) bool     { return false }
func (testEnvProvider) MustGet(_ string) string { return "" }

type testLogProvider struct{}

func (testLogProvider) Debug(_ string, _ map[string]any)    {}
func (testLogProvider) Info(_ string, _ map[string]any)     {}
func (testLogProvider) Warning(_ string, _ map[string]any)  {}
func (testLogProvider) Error(_ string, _ map[string]any)    {}
func (testLogProvider) Critical(_ string, _ map[string]any) {}
func (testLogProvider) SetLogLevel(_ logdomain.LogLevel)    {}

func newTestEchoAPIProvider() *apidriver.EchoApiProvider {
	hub := &providerdomain.ProviderHub{}
	providerdomain.Register[envdomain.EnvProvider](hub, testEnvProvider{})
	providerdomain.Register[logdomain.LogProvider](hub, testLogProvider{})

	return apidriver.NewEchoApiProvider("8080", hub)
}

func handlerPointer(handler echo.HTTPErrorHandler) uintptr {
	return reflect.ValueOf(handler).Pointer()
}

// Package modules_test verifies the route wiring provided by Raidark's core
// Echo modules without invoking downstream business controllers.
package modules_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/driver/modules"
	authdomain "github.com/r0x16/Raidark/shared/auth/domain"
	providerdomain "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoMainModule_RegistersBaseRoutes(t *testing.T) {
	hub, apiProvider := newMetricsModuleTestHub()
	module := &modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}

	require.NoError(t, module.Setup())

	assertRouteRegistered(t, apiProvider.Server, http.MethodGet, "/health")
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	apiProvider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "OK", recorder.Body.String())
}

func TestEchoAuthModule_RegistersAuthEndpoints(t *testing.T) {
	hub, apiProvider := newMetricsModuleTestHub()
	module := &modules.EchoAuthModule{EchoModule: modules.NewEchoModule("/auth", hub)}

	require.NoError(t, module.Setup())

	assertRouteRegistered(t, apiProvider.Server, http.MethodPost, "/auth/exchange")
	assertRouteRegistered(t, apiProvider.Server, http.MethodPost, "/auth/refresh")
	assertRouteRegistered(t, apiProvider.Server, http.MethodPost, "/auth/logout")
	assert.Len(t, module.GetModel(), 1)
}

func TestEchoApiMainModule_RegistersPingAndAuthenticatedMe(t *testing.T) {
	hub, apiProvider := newMetricsModuleTestHub()
	module := &modules.EchoApiMainModule{EchoModule: modules.NewEchoModule("/api", hub)}

	require.NoError(t, module.Setup())

	assertRouteRegistered(t, apiProvider.Server, http.MethodGet, "/api/ping")
	assertRouteRegistered(t, apiProvider.Server, http.MethodGet, "/api/me")

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	apiProvider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "pong", recorder.Body.String())
}

func TestEchoApiMainModule_ReturnsClaimsFromContext(t *testing.T) {
	hub, apiProvider := newMetricsModuleTestHub()
	module := &modules.EchoApiMainModule{EchoModule: modules.NewEchoModule("/api", hub)}
	module.Group.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("user", &authdomain.Claims{Username: "alice"})
			return next(c)
		}
	})

	require.NoError(t, module.Setup())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	apiProvider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]any
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.Equal(t, "alice", body["Username"])
}

func TestEchoModuleActionInjection_PanicsWhenHubIsMissing(t *testing.T) {
	module := &modules.EchoModule{}

	assert.PanicsWithValue(t, "Hub is not set in EchoModule", func() {
		_ = module.ActionInjection(func(echo.Context, *providerdomain.ProviderHub) error { return nil })
	})
}

func TestNewEchoModule_PanicsWhenApiProviderIsMissing(t *testing.T) {
	hub := &providerdomain.ProviderHub{}

	assert.PanicsWithError(t, "provider *reflect.rtype not found", func() {
		_ = modules.NewEchoModule("", hub)
	})
}

func assertRouteRegistered(t *testing.T, server *echo.Echo, method string, path string) {
	t.Helper()
	_ = findRoute(t, server, method, path)
}

func findRoute(t *testing.T, server *echo.Echo, method string, path string) *echo.Route {
	t.Helper()
	for _, route := range server.Routes() {
		if route.Method == method && route.Path == path {
			return route
		}
	}
	t.Fatalf("route %s %s was not registered", method, path)
	return nil
}

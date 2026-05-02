// Package driver_test verifies the runtime CORS and CSRF policy toggles exposed
// by the shared Echo provider.
package driver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/driver/modules"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoApiProvider_CSRFDisabledLeavesPOSTOpenAndTokenRouteMissing(t *testing.T) {
	hub, provider := newCoreAPIProvider(t, coreEnvProvider{
		boolValues: map[string]bool{"CSRF_ENABLED": false},
	})

	require.NoError(t, provider.Setup())
	require.NoError(t, (&modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}).Setup())
	provider.Server.POST("/protected", noContentHandler(http.StatusNoContent))

	postRecorder := httptest.NewRecorder()
	postRequest := httptest.NewRequest(http.MethodPost, "/protected", nil)
	provider.Server.ServeHTTP(postRecorder, postRequest)

	require.Equal(t, http.StatusNoContent, postRecorder.Code)

	tokenRecorder := httptest.NewRecorder()
	tokenRequest := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	provider.Server.ServeHTTP(tokenRecorder, tokenRequest)

	assertRouteMissing(t, provider.Server.Routes(), http.MethodGet, "/csrf-token")
}

func TestEchoApiProvider_CSRFDefaultsToDisabled(t *testing.T) {
	hub, provider := newCoreAPIProvider(t, coreEnvProvider{})

	require.NoError(t, provider.Setup())
	require.NoError(t, (&modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}).Setup())
	provider.Server.POST("/protected", noContentHandler(http.StatusNoContent))

	postRecorder := httptest.NewRecorder()
	postRequest := httptest.NewRequest(http.MethodPost, "/protected", nil)
	provider.Server.ServeHTTP(postRecorder, postRequest)

	assert.Equal(t, http.StatusNoContent, postRecorder.Code)
}

func TestEchoApiProvider_CSRFEnabledRejectsPOSTWithoutToken(t *testing.T) {
	hub, provider := newCoreAPIProvider(t, csrfEnabledEnv())

	require.NoError(t, provider.Setup())
	require.NoError(t, (&modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}).Setup())
	provider.Server.POST("/protected", noContentHandler(http.StatusNoContent))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/protected", nil)
	request.Header.Set("X-Correlation-ID", "trace-rdk-004-csrf")
	provider.Server.ServeHTTP(recorder, request)

	require.GreaterOrEqual(t, recorder.Code, http.StatusBadRequest)
	assert.Contains(t, recorder.Body.String(), `"error"`)
	assert.Contains(t, recorder.Body.String(), `"trace_id":"trace-rdk-004-csrf"`)
}

func TestEchoApiProvider_CSRFEnabledAcceptsTokenFromEndpoint(t *testing.T) {
	hub, provider := newCoreAPIProvider(t, csrfEnabledEnv())

	require.NoError(t, provider.Setup())
	require.NoError(t, (&modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}).Setup())
	provider.Server.POST("/protected", noContentHandler(http.StatusNoContent))

	tokenRecorder := httptest.NewRecorder()
	tokenRequest := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	provider.Server.ServeHTTP(tokenRecorder, tokenRequest)

	require.Equal(t, http.StatusOK, tokenRecorder.Code)
	csrfToken := decodeCSRFToken(t, tokenRecorder.Body.Bytes())

	postRecorder := httptest.NewRecorder()
	postRequest := httptest.NewRequest(http.MethodPost, "/protected", nil)
	postRequest.Header.Set("X-CSRF-Token", csrfToken)
	for _, cookie := range tokenRecorder.Result().Cookies() {
		postRequest.AddCookie(cookie)
	}
	provider.Server.ServeHTTP(postRecorder, postRequest)

	assert.Equal(t, http.StatusNoContent, postRecorder.Code)
}

func TestEchoApiProvider_CORSUnsetDoesNotMountMiddleware(t *testing.T) {
	_, provider := newCoreAPIProvider(t, coreEnvProvider{})

	require.NoError(t, provider.Setup())
	provider.Server.GET("/protected", noContentHandler(http.StatusOK))

	recorder := httptest.NewRecorder()
	request := newPreflightRequest("/protected", "https://a.example", http.MethodGet)
	provider.Server.ServeHTTP(recorder, request)

	assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Methods"))
}

func TestEchoApiProvider_CORSAllowsConfiguredOrigins(t *testing.T) {
	_, provider := newCoreAPIProvider(t, corsEnabledEnv())

	require.NoError(t, provider.Setup())
	provider.Server.GET("/protected", noContentHandler(http.StatusOK))

	for _, origin := range []string{"https://a.example", "https://b.example"} {
		t.Run(origin, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := newPreflightRequest("/protected", origin, http.MethodGet)
			provider.Server.ServeHTTP(recorder, request)

			require.Equal(t, http.StatusNoContent, recorder.Code)
			assert.Equal(t, origin, recorder.Header().Get("Access-Control-Allow-Origin"))
		})
	}

	recorder := httptest.NewRecorder()
	request := newPreflightRequest("/protected", "https://c.example", http.MethodGet)
	provider.Server.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Empty(t, recorder.Header().Get("Access-Control-Allow-Origin"))
}

func TestEchoApiProvider_CORSRespectsHeadersMethodsCredentialsAndMaxAge(t *testing.T) {
	_, provider := newCoreAPIProvider(t, corsEnabledEnv())

	require.NoError(t, provider.Setup())
	provider.Server.PATCH("/protected", noContentHandler(http.StatusNoContent))

	recorder := httptest.NewRecorder()
	request := newPreflightRequest("/protected", "https://a.example", http.MethodPatch)
	request.Header.Set("Access-Control-Request-Headers", "Authorization,X-Correlation-ID")
	provider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusNoContent, recorder.Code)
	assert.Equal(t, "https://a.example", recorder.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", recorder.Header().Get("Access-Control-Allow-Credentials"))
	assert.Equal(t, "600", recorder.Header().Get("Access-Control-Max-Age"))
	assertHeaderListContains(t, recorder.Header().Get("Access-Control-Allow-Methods"), http.MethodPatch)
	assertHeaderListContains(t, recorder.Header().Get("Access-Control-Allow-Headers"), "Authorization")
	assertHeaderListContains(t, recorder.Header().Get("Access-Control-Allow-Headers"), "X-Correlation-ID")
}

func csrfEnabledEnv() coreEnvProvider {
	return coreEnvProvider{
		boolValues: map[string]bool{"CSRF_ENABLED": true},
		intValues: map[string]int{
			"CSRF_TOKEN_LENGTH":   32,
			"CSRF_COOKIE_MAX_AGE": 86400,
		},
		strings: map[string]string{
			"CSRF_TOKEN_LOOKUP": "header:X-CSRF-Token",
		},
	}
}

func corsEnabledEnv() coreEnvProvider {
	return coreEnvProvider{
		boolValues: map[string]bool{"CORS_ALLOW_CREDENTIALS": true},
		intValues:  map[string]int{"CORS_MAX_AGE": 600},
		sliceSet:   map[string]bool{"CORS_ALLOW_ORIGINS": true},
		slices: map[string][]string{
			"CORS_ALLOW_ORIGINS": []string{"https://a.example", "https://b.example"},
			"CORS_ALLOW_HEADERS": []string{"Content-Type", "Authorization", "X-Correlation-ID"},
			"CORS_ALLOW_METHODS": []string{http.MethodGet, http.MethodPatch, http.MethodOptions},
		},
	}
}

func noContentHandler(status int) func(c echo.Context) error {
	return func(c echo.Context) error {
		return c.NoContent(status)
	}
}

func newPreflightRequest(target string, origin string, method string) *http.Request {
	request := httptest.NewRequest(http.MethodOptions, target, nil)
	request.Header.Set("Origin", origin)
	request.Header.Set("Access-Control-Request-Method", method)
	return request
}

func decodeCSRFToken(t *testing.T, body []byte) string {
	t.Helper()
	var payload map[string]string
	require.NoError(t, json.Unmarshal(body, &payload))
	require.NotEmpty(t, payload["csrf_token"])
	return payload["csrf_token"]
}

func assertRouteMissing(t *testing.T, routes []*echo.Route, method string, path string) {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			t.Fatalf("route %s %s should not be registered", method, path)
		}
	}
}

func assertHeaderListContains(t *testing.T, values string, want string) {
	t.Helper()
	for _, value := range strings.Split(values, ",") {
		if strings.TrimSpace(value) == want {
			return
		}
	}
	t.Fatalf("header list %q does not contain %q", values, want)
}

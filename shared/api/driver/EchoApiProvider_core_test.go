// Package driver_test covers core HTTP provider behavior that every Raidark
// service inherits from the shared Echo adapter.
package driver_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	apidriver "github.com/r0x16/Raidark/shared/api/driver"
	"github.com/r0x16/Raidark/shared/api/driver/modules"
	envdomain "github.com/r0x16/Raidark/shared/env/domain"
	eventdomain "github.com/r0x16/Raidark/shared/events/domain"
	logdomain "github.com/r0x16/Raidark/shared/logger/domain"
	providerdomain "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoApiProvider_BootstrapsMainModuleAndHealthCheck(t *testing.T) {
	hub, provider := newCoreAPIProvider(t, coreEnvProvider{})
	mainModule := &modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}

	require.NoError(t, provider.Setup())
	require.NoError(t, mainModule.Setup())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	provider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "OK", recorder.Body.String())
	assert.NotNil(t, provider.Server)
}

func TestEchoApiProvider_MountsCSRFTokenRouteWhenEnabled(t *testing.T) {
	hub, provider := newCoreAPIProvider(t, coreEnvProvider{
		boolValues: map[string]bool{"CSRF_ENABLED": true},
		intValues:  map[string]int{"CSRF_TOKEN_LENGTH": 32},
	})
	mainModule := &modules.EchoMainModule{EchoModule: modules.NewEchoModule("", hub)}

	require.NoError(t, provider.Setup())
	require.NoError(t, mainModule.Setup())

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/csrf-token", nil)
	provider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	var body map[string]string
	require.NoError(t, json.Unmarshal(recorder.Body.Bytes(), &body))
	assert.NotEmpty(t, body["csrf_token"])
}

func TestEchoApiProvider_RegisterKeepsCustomModulesInOrder(t *testing.T) {
	_, provider := newCoreAPIProvider(t, coreEnvProvider{})
	first := namedAPIModule{name: "first"}
	second := namedAPIModule{name: "second"}

	provider.Register(first)
	provider.Register(second)

	require.Len(t, provider.ProvidesModules(), 2)
	assert.Equal(t, "first", provider.ProvidesModules()[0].Name())
	assert.Equal(t, "second", provider.ProvidesModules()[1].Name())
}

func TestApplicationBundle_ActionInjectionPassesBundleToHandler(t *testing.T) {
	bundle := &apidriver.ApplicationBundle{Env: coreEnvProvider{}}
	handler := bundle.ActionInjection(func(c echo.Context, received *apidriver.ApplicationBundle) error {
		require.Same(t, bundle, received)
		return c.String(http.StatusAccepted, "bundle injected")
	})

	e := echo.New()
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/bundle", nil)

	require.NoError(t, handler(e.NewContext(request, recorder)))
	assert.Equal(t, http.StatusAccepted, recorder.Code)
	assert.Equal(t, "bundle injected", recorder.Body.String())
}

func newCoreAPIProvider(t *testing.T, env coreEnvProvider) (*providerdomain.ProviderHub, *apidriver.EchoApiProvider) {
	t.Helper()

	hub := &providerdomain.ProviderHub{}
	providerdomain.Register[envdomain.EnvProvider](hub, env)
	providerdomain.Register[logdomain.LogProvider](hub, testLogProvider{})
	provider := apidriver.NewEchoApiProvider("8080", hub)
	providerdomain.Register[apidomain.ApiProvider](hub, provider)
	return hub, provider
}

type coreEnvProvider struct {
	boolValues map[string]bool
	intValues  map[string]int
	sliceSet   map[string]bool
	slices     map[string][]string
	strings    map[string]string
}

func (p coreEnvProvider) GetString(key string, defaultValue string) string {
	if value, ok := p.strings[key]; ok {
		return value
	}
	return defaultValue
}

func (p coreEnvProvider) GetBool(key string, defaultValue bool) bool {
	if value, ok := p.boolValues[key]; ok {
		return value
	}
	return defaultValue
}

func (p coreEnvProvider) GetInt(key string, defaultValue int) int {
	if value, ok := p.intValues[key]; ok {
		return value
	}
	return defaultValue
}

func (p coreEnvProvider) GetFloat(_ string, defaultValue float64) float64 {
	return defaultValue
}

func (p coreEnvProvider) GetSlice(key string, defaultValue []string) []string {
	if value, ok := p.slices[key]; ok {
		return value
	}
	return defaultValue
}

func (p coreEnvProvider) GetSliceWithSeparator(key string, _ string, defaultValue []string) []string {
	return p.GetSlice(key, defaultValue)
}

func (p coreEnvProvider) IsSet(key string) bool {
	return p.sliceSet[key]
}

func (p coreEnvProvider) MustGet(key string) string {
	return p.strings[key]
}

type namedAPIModule struct {
	name string
}

func (m namedAPIModule) Name() string                                 { return m.name }
func (namedAPIModule) Setup() error                                   { return nil }
func (namedAPIModule) GetModel() []any                                { return nil }
func (namedAPIModule) GetSeedData() []any                             { return nil }
func (namedAPIModule) GetEventListeners() []eventdomain.EventListener { return nil }

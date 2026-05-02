// Package modules_test verifies API module wiring that mounts optional
// observability endpoints.
package modules_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	apidriver "github.com/r0x16/Raidark/shared/api/driver"
	"github.com/r0x16/Raidark/shared/api/driver/modules"
	envdomain "github.com/r0x16/Raidark/shared/env/domain"
	logdomain "github.com/r0x16/Raidark/shared/logger/domain"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	obsdriver "github.com/r0x16/Raidark/shared/observability/driver"
	providerdomain "github.com/r0x16/Raidark/shared/providers/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEchoMetricsModule_MountsMetricsRouteWhenProviderExists(t *testing.T) {
	hub, apiProvider := newMetricsModuleTestHub()
	metricsProvider := obsdriver.NewPrometheusMetricsProvider("/internal/metrics")
	metricsProvider.Metrics().SetOutboxPending(3)
	providerdomain.Register[obsdomain.MetricsProvider](hub, metricsProvider)
	module := &modules.EchoMetricsModule{EchoModule: modules.NewEchoModule("", hub)}

	require.NoError(t, module.Setup())

	request := httptest.NewRequest(http.MethodGet, "/internal/metrics", nil)
	recorder := httptest.NewRecorder()
	apiProvider.Server.ServeHTTP(recorder, request)

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "outbox_pending_gauge")
	assert.Contains(t, recorder.Body.String(), " 3")
}

func TestEchoMetricsModule_IsNoopWithoutMetricsProvider(t *testing.T) {
	hub, apiProvider := newMetricsModuleTestHub()
	module := &modules.EchoMetricsModule{EchoModule: modules.NewEchoModule("", hub)}

	require.NoError(t, module.Setup())

	request := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	recorder := httptest.NewRecorder()
	apiProvider.Server.ServeHTTP(recorder, request)

	assert.Equal(t, http.StatusNotFound, recorder.Code)
}

func newMetricsModuleTestHub() (*providerdomain.ProviderHub, *apidriver.EchoApiProvider) {
	hub := &providerdomain.ProviderHub{}
	providerdomain.Register[envdomain.EnvProvider](hub, moduleEnvProvider{})
	providerdomain.Register[logdomain.LogProvider](hub, moduleLogProvider{})
	apiProvider := apidriver.NewEchoApiProvider("8080", hub)
	providerdomain.Register[apidomain.ApiProvider](hub, apiProvider)
	return hub, apiProvider
}

type moduleEnvProvider struct{}

func (moduleEnvProvider) GetString(_ string, defaultValue string) string { return defaultValue }
func (moduleEnvProvider) GetBool(_ string, defaultValue bool) bool       { return defaultValue }
func (moduleEnvProvider) GetInt(_ string, defaultValue int) int          { return defaultValue }
func (moduleEnvProvider) GetFloat(_ string, defaultValue float64) float64 {
	return defaultValue
}
func (moduleEnvProvider) GetSlice(_ string, defaultValue []string) []string {
	return defaultValue
}
func (moduleEnvProvider) GetSliceWithSeparator(_ string, _ string, defaultValue []string) []string {
	return defaultValue
}
func (moduleEnvProvider) IsSet(_ string) bool     { return false }
func (moduleEnvProvider) MustGet(_ string) string { return "" }

type moduleLogProvider struct{}

func (moduleLogProvider) Debug(_ string, _ map[string]any)    {}
func (moduleLogProvider) Info(_ string, _ map[string]any)     {}
func (moduleLogProvider) Warning(_ string, _ map[string]any)  {}
func (moduleLogProvider) Error(_ string, _ map[string]any)    {}
func (moduleLogProvider) Critical(_ string, _ map[string]any) {}
func (moduleLogProvider) SetLogLevel(_ logdomain.LogLevel)    {}

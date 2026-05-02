// Package driver_test verifies provider-factory wiring visible through the
// shared provider hub.
package driver_test

import (
	"testing"

	envdomain "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/observability"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	providerdomain "github.com/r0x16/Raidark/shared/providers/domain"
	providerdriver "github.com/r0x16/Raidark/shared/providers/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsProviderFactory_RegistersProviderWhenMetricsEnabled(t *testing.T) {
	hub := &providerdomain.ProviderHub{}
	providerdomain.Register[envdomain.EnvProvider](hub, mapEnvProvider{
		strings: map[string]string{
			"METRICS_PATH": "/internal/metrics",
			"SERVICE_NAME": "orders",
		},
		bools: map[string]bool{"METRICS_ENABLED": true},
	})
	factory := &providerdriver.MetricsProviderFactory{}
	factory.Init(hub)

	require.NoError(t, factory.Register(hub))

	require.True(t, providerdomain.Exists[obsdomain.MetricsProvider](hub))
	provider := providerdomain.Get[obsdomain.MetricsProvider](hub)
	assert.Equal(t, "/internal/metrics", provider.Path())
	assert.Equal(t, "orders", observability.GetDefaultServiceName())
}

func TestMetricsProviderFactory_SkipsProviderWhenMetricsDisabled(t *testing.T) {
	hub := &providerdomain.ProviderHub{}
	providerdomain.Register[envdomain.EnvProvider](hub, mapEnvProvider{
		strings: map[string]string{"SERVICE_NAME": "workers"},
		bools:   map[string]bool{"METRICS_ENABLED": false},
	})
	factory := &providerdriver.MetricsProviderFactory{}
	factory.Init(hub)

	require.NoError(t, factory.Register(hub))

	assert.False(t, providerdomain.Exists[obsdomain.MetricsProvider](hub))
	assert.Equal(t, "workers", observability.GetDefaultServiceName())
}

type mapEnvProvider struct {
	strings map[string]string
	bools   map[string]bool
}

func (m mapEnvProvider) GetString(key, defaultValue string) string {
	if value, ok := m.strings[key]; ok {
		return value
	}
	return defaultValue
}

func (m mapEnvProvider) GetBool(key string, defaultValue bool) bool {
	if value, ok := m.bools[key]; ok {
		return value
	}
	return defaultValue
}

func (m mapEnvProvider) GetInt(_ string, defaultValue int) int { return defaultValue }
func (m mapEnvProvider) GetFloat(_ string, defaultValue float64) float64 {
	return defaultValue
}
func (m mapEnvProvider) GetSlice(_ string, defaultValue []string) []string {
	return defaultValue
}
func (m mapEnvProvider) GetSliceWithSeparator(_ string, _ string, defaultValue []string) []string {
	return defaultValue
}
func (m mapEnvProvider) IsSet(key string) bool {
	_, ok := m.strings[key]
	return ok
}
func (m mapEnvProvider) MustGet(key string) string {
	if value, ok := m.strings[key]; ok {
		return value
	}
	panic("missing env: " + key)
}

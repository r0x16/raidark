package driver

import (
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/observability"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
	obsdriver "github.com/r0x16/Raidark/shared/observability/driver"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

// MetricsProviderFactory wires the Prometheus-backed MetricsProvider into
// the shared provider hub when METRICS_ENABLED=true. Adding this factory to
// main.go's providers list is the explicit opt-in: services that don't add
// it never construct a Metrics bundle and never expose /metrics.
//
// Even when the factory is in the providers list, it skips registration
// when METRICS_ENABLED=false — so ops can flip metrics off via env vars
// without rebuilding the binary. Consumers (EchoApiProvider, the
// EchoMetricsModule) probe the hub with domprovider.Exists before using
// the provider, so the absence is harmless.
type MetricsProviderFactory struct {
	env domenv.EnvProvider
}

// Init implements ProviderFactory. The factory needs the env provider to
// resolve METRICS_ENABLED, METRICS_PATH and SERVICE_NAME, so it captures a
// typed reference at this stage.
func (f *MetricsProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

// Register builds the provider and stores it on the hub when metrics are
// enabled. SERVICE_NAME is registered globally regardless of metrics state
// because it is also consumed by the observability logger to stamp the
// service field on every log line.
func (f *MetricsProviderFactory) Register(hub *domain.ProviderHub) error {
	// Register SERVICE_NAME first; the logger reads it via the global
	// default even when metrics are disabled.
	if name := f.env.GetString("SERVICE_NAME", ""); name != "" {
		observability.SetDefaultServiceName(name)
	}

	if !f.env.GetBool("METRICS_ENABLED", true) {
		return nil
	}

	path := f.env.GetString("METRICS_PATH", "/metrics")
	provider := obsdriver.NewPrometheusMetricsProvider(path)
	domain.Register[obsdomain.MetricsProvider](hub, provider)
	return nil
}

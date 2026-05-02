package driver

import (
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/observability"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

// MetricsProviderFactory wires the Prometheus-backed MetricsProvider into the
// shared provider hub. It reads METRICS_ENABLED to decide whether the
// scrape endpoint should be mounted; the underlying collectors are always
// constructed so that record-side calls (counter increments, histogram
// observations) are nil-safe even when scraping is off.
type MetricsProviderFactory struct {
	env domenv.EnvProvider
}

// Init implements ProviderFactory. The factory needs the env provider to
// resolve METRICS_ENABLED, so it captures a typed reference at this stage.
func (f *MetricsProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

// Register builds the provider and stores it on the hub under the
// observability.MetricsProvider interface key, so consumers stay loosely
// coupled to the Prometheus implementation.
func (f *MetricsProviderFactory) Register(hub *domain.ProviderHub) error {
	enabled := f.env.GetBool("METRICS_ENABLED", true)
	provider := observability.NewPrometheusMetricsProvider(enabled)

	// Register the SERVICE_NAME so log.FromContext can stamp it into every
	// log line emitted by goroutines that don't carry the request context.
	if name := f.env.GetString("SERVICE_NAME", ""); name != "" {
		observability.SetDefaultServiceName(name)
	}

	domain.Register[observability.MetricsProvider](hub, provider)
	return nil
}

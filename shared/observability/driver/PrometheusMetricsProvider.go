// Package driver contains the concrete observability adapters that bind the
// domain contracts (interfaces under shared/observability/domain) to a
// specific runtime — currently Prometheus for metrics scraping. New adapters
// (OTLP, statsd, …) belong here so the rest of the codebase only depends on
// the framework-neutral domain layer.
package driver

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/r0x16/Raidark/shared/observability"
	obsdomain "github.com/r0x16/Raidark/shared/observability/domain"
)

// PrometheusMetricsProvider is the default MetricsProvider. It owns the
// canonical observability.Metrics bundle plus a private Prometheus
// registry, and exposes both as required by the domain contract.
//
// One provider per process is the typical pattern. Tests should construct
// their own provider rather than mutating a shared one so cardinality and
// counter values do not leak across cases.
type PrometheusMetricsProvider struct {
	metrics *observability.Metrics
	path    string
}

var _ obsdomain.MetricsProvider = (*PrometheusMetricsProvider)(nil)

// NewPrometheusMetricsProvider constructs the provider with a fresh metrics
// bundle and the URL path under which the scrape endpoint will be exposed.
// The factory is a separate concern — it decides whether to instantiate the
// provider at all, based on METRICS_ENABLED.
func NewPrometheusMetricsProvider(path string) *PrometheusMetricsProvider {
	return &PrometheusMetricsProvider{
		metrics: observability.NewMetrics(),
		path:    path,
	}
}

// Metrics implements MetricsProvider.
func (p *PrometheusMetricsProvider) Metrics() *observability.Metrics {
	return p.metrics
}

// Path implements MetricsProvider.
func (p *PrometheusMetricsProvider) Path() string {
	return p.path
}

// Handler implements MetricsProvider. We use promhttp.HandlerFor against
// the provider's private registry rather than promhttp.Handler() (which
// uses the default registry) so tests can run multiple providers in
// parallel without colliding on collector names.
func (p *PrometheusMetricsProvider) Handler() http.Handler {
	return promhttp.HandlerFor(p.metrics.Registry, promhttp.HandlerOpts{
		Registry: p.metrics.Registry,
	})
}

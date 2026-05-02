package observability

import (
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// MetricsProvider exposes the Prometheus collectors plus the helper that
// mounts the scrape endpoint on an Echo server. The interface lives next to
// the implementation (rather than in a domain package) because there is one
// natural implementation — Prometheus — and indirection layers would only
// add ceremony without buying flexibility.
type MetricsProvider interface {
	// Metrics returns the underlying Metrics bundle so middlewares and
	// publisher/consumer adapters can record onto the registered collectors.
	Metrics() *Metrics

	// MountScrapeEndpoint registers a GET handler at path on srv that
	// serves the Prometheus exposition format. Idempotent: callers are
	// expected to invoke it exactly once per server.
	MountScrapeEndpoint(srv *echo.Echo, path string)

	// Enabled reports whether metrics collection is turned on for this
	// process. When false, MountScrapeEndpoint is a no-op and the
	// Metrics bundle is still safe to use (writes are dropped because no
	// one scrapes them, but the collectors don't blow up).
	Enabled() bool
}

// PrometheusMetricsProvider is the default MetricsProvider. It is constructed
// at boot from env vars and registered on the provider hub so middlewares
// and event drivers can pull it out via domprovider.Get.
type PrometheusMetricsProvider struct {
	metrics *Metrics
	enabled bool
}

var _ MetricsProvider = (*PrometheusMetricsProvider)(nil)

// NewPrometheusMetricsProvider builds the provider. The caller decides
// whether to enable scraping (typically driven by METRICS_ENABLED). When
// enabled is false the provider still owns a Metrics bundle so call sites
// don't need nil checks; only the scrape endpoint stays unregistered.
func NewPrometheusMetricsProvider(enabled bool) *PrometheusMetricsProvider {
	return &PrometheusMetricsProvider{
		metrics: NewMetrics(),
		enabled: enabled,
	}
}

// Metrics implements MetricsProvider.
func (p *PrometheusMetricsProvider) Metrics() *Metrics { return p.metrics }

// Enabled implements MetricsProvider.
func (p *PrometheusMetricsProvider) Enabled() bool { return p.enabled }

// MountScrapeEndpoint implements MetricsProvider. We use promhttp.HandlerFor
// against the provider's private registry rather than promhttp.Handler()
// (which uses the default registry) so tests can run multiple providers in
// parallel without colliding on collector names.
func (p *PrometheusMetricsProvider) MountScrapeEndpoint(srv *echo.Echo, path string) {
	if !p.enabled || srv == nil || path == "" {
		return
	}
	handler := promhttp.HandlerFor(p.metrics.Registry, promhttp.HandlerOpts{
		Registry: p.metrics.Registry,
	})
	srv.GET(path, echo.WrapHandler(handler))
}

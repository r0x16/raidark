// Package domain holds the framework-agnostic contracts of the observability
// layer. Anything that other packages depend on for type-checking lives
// here; concrete adapters (Prometheus, OTLP, …) live under
// shared/observability/driver.
package domain

import (
	"net/http"

	"github.com/r0x16/Raidark/shared/observability"
)

// MetricsProvider is the cross-cutting contract for metric collection. It
// gives middlewares and event publishers/consumers access to the canonical
// metrics bundle and exposes the scrape endpoint as a plain http.Handler
// so the wire format stays decoupled from the HTTP framework. Code that
// needs to mount the scrape endpoint on Echo (or any other router) does so
// via the dedicated ApiModule under shared/observability/driver, not by
// passing an *echo.Echo into this interface.
type MetricsProvider interface {
	// Metrics returns the underlying Metrics bundle so callers can
	// increment counters, observe histograms, etc., directly.
	Metrics() *observability.Metrics

	// Handler returns the http.Handler that serves the scrape endpoint
	// in Prometheus exposition format. The handler reads from the
	// provider's private registry, so multiple providers in the same
	// process do not collide.
	Handler() http.Handler

	// Path returns the URL path under which the scrape endpoint should
	// be mounted (default "/metrics", overridable by env). Centralising
	// the path here means both the ApiModule and any health-check page
	// that links to /metrics agree on the same value.
	Path() string
}

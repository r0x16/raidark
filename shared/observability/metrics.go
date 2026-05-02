package observability

import (
	"github.com/prometheus/client_golang/prometheus"
)

// HTTP histogram buckets in milliseconds. The bucket layout is intentionally
// shallow (six buckets) to keep cardinality of /metrics low while still
// covering the latency regimes we care about: fast hits (<25ms), interactive
// (<500ms), slow (<5s), and timeout-territory (>5s lands in +Inf).
var defaultHTTPDurationBuckets = []float64{5, 25, 100, 500, 1000, 5000}

// Event processing histogram buckets in milliseconds. Wider top end than
// HTTP because batch consumers often legitimately take seconds to handle a
// single message (DB writes, webhook fan-outs).
var defaultEventDurationBuckets = []float64{5, 25, 100, 500, 1000, 5000, 30000}

// Metrics is the registry plus the canonical collectors used across Raidark.
// Construct one per process via NewMetrics; pass it to middlewares and event
// publishers/consumers. Tests should construct a private Metrics instance
// rather than mutating a global so cardinality and counter values don't
// leak across cases.
type Metrics struct {
	// Registry is the Prometheus registry that hosts every collector below.
	// Exported so adapter code (e.g. /metrics handler) can scrape it without
	// going through the global default registry, which keeps tests isolated
	// and avoids accidental collisions with library-registered collectors.
	Registry *prometheus.Registry

	// HTTPRequestsTotal counts every HTTP request handled by the server,
	// labelled by status code and matched route pattern (Echo's c.Path()),
	// not the raw URL — using the URL would explode cardinality on routes
	// with path parameters.
	HTTPRequestsTotal *prometheus.CounterVec

	// HTTPRequestDurationMs is the per-route latency histogram in
	// milliseconds. The "endpoint" label uses the same matched route
	// pattern as HTTPRequestsTotal so the two can be joined safely.
	HTTPRequestDurationMs *prometheus.HistogramVec

	// EventsPublishedTotal counts publish attempts with an "outcome" label
	// (success, failure, etc.) so dashboards can compute a publish error
	// rate without joining counters.
	EventsPublishedTotal *prometheus.CounterVec

	// EventsConsumedTotal counts consume attempts. Includes a "consumer"
	// label because the same subject is typically delivered to several
	// independent consumers and per-consumer health is what oncall watches.
	EventsConsumedTotal *prometheus.CounterVec

	// EventsRedeliveriesTotal counts how often a message had to be
	// redelivered to a consumer. Spikes here signal poison pills, slow
	// consumers, or downstream outages.
	EventsRedeliveriesTotal *prometheus.CounterVec

	// EventProcessingDurationMs is the consumer-side processing latency
	// histogram. Useful both for SLOs and for spotting consumers that
	// regress when downstream dependencies degrade.
	EventProcessingDurationMs *prometheus.HistogramVec

	// OutboxPending exposes the current depth of the transactional outbox.
	// A monotonically rising value usually indicates the outbox publisher
	// has fallen behind and is the most direct signal for outbox-related
	// alerting.
	OutboxPending prometheus.Gauge
}

// NewMetrics constructs a Metrics bundle with all collectors registered on a
// fresh registry. It is safe to construct multiple bundles per process — they
// are independent — but the typical pattern is one per process, registered in
// the provider hub at boot.
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		Registry: registry,

		HTTPRequestsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests handled, labelled by HTTP status code and matched route pattern.",
			},
			[]string{"status", "endpoint"},
		),

		HTTPRequestDurationMs: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_ms",
				Help:    "HTTP request duration in milliseconds, by matched route pattern.",
				Buckets: defaultHTTPDurationBuckets,
			},
			[]string{"endpoint"},
		),

		EventsPublishedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "events_published_total",
				Help: "Total number of domain events published, labelled by subject and outcome.",
			},
			[]string{"subject", "outcome"},
		),

		EventsConsumedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "events_consumed_total",
				Help: "Total number of domain events consumed, labelled by subject, consumer and outcome.",
			},
			[]string{"subject", "consumer", "outcome"},
		),

		EventsRedeliveriesTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "events_redeliveries_total",
				Help: "Total number of message redeliveries observed by the consumer side.",
			},
			[]string{"subject", "consumer"},
		),

		EventProcessingDurationMs: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "event_processing_duration_ms",
				Help:    "Event consumer processing duration in milliseconds.",
				Buckets: defaultEventDurationBuckets,
			},
			[]string{"subject", "consumer"},
		),

		OutboxPending: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "outbox_pending_gauge",
				Help: "Current depth of the transactional outbox awaiting publication.",
			},
		),
	}

	// MustRegister panics on collision; collisions can only happen here if
	// someone reuses the same registry — which we never do for Metrics —
	// so a panic at boot is the desired loud-fail behaviour.
	registry.MustRegister(
		m.HTTPRequestsTotal,
		m.HTTPRequestDurationMs,
		m.EventsPublishedTotal,
		m.EventsConsumedTotal,
		m.EventsRedeliveriesTotal,
		m.EventProcessingDurationMs,
		m.OutboxPending,
	)

	return m
}

// RecordEventPublished is a thin sugar over the underlying CounterVec for
// publishers. Centralising the label order here means call sites cannot pass
// the labels in the wrong slot and silently corrupt the time series.
func (m *Metrics) RecordEventPublished(subject, outcome string) {
	m.EventsPublishedTotal.WithLabelValues(subject, outcome).Inc()
}

// RecordEventConsumed is the consumer-side counterpart to
// RecordEventPublished. Same rationale: encapsulate label order.
func (m *Metrics) RecordEventConsumed(subject, consumer, outcome string) {
	m.EventsConsumedTotal.WithLabelValues(subject, consumer, outcome).Inc()
}

// RecordEventRedelivery increments the per-(subject, consumer) redelivery
// counter. Call this when the message broker reports a redelivery, not on
// every retry inside a single delivery — the two are different signals.
func (m *Metrics) RecordEventRedelivery(subject, consumer string) {
	m.EventsRedeliveriesTotal.WithLabelValues(subject, consumer).Inc()
}

// ObserveEventProcessing records a consumer-side processing duration in
// milliseconds. The float64 type matches Prometheus's Observe contract.
func (m *Metrics) ObserveEventProcessing(subject, consumer string, durationMs float64) {
	m.EventProcessingDurationMs.WithLabelValues(subject, consumer).Observe(durationMs)
}

// SetOutboxPending sets the outbox depth gauge. Typically called by the
// outbox publisher loop once per polling iteration with the SELECT COUNT(*)
// of unpublished rows.
func (m *Metrics) SetOutboxPending(pending float64) {
	m.OutboxPending.Set(pending)
}

# Prometheus metrics

Packages:
- `github.com/r0x16/Raidark/shared/observability` — `Metrics` bundle, `HTTPMetrics` middleware
- `github.com/r0x16/Raidark/shared/observability/domain` — `MetricsProvider` interface
- `github.com/r0x16/Raidark/shared/observability/driver` — `PrometheusMetricsProvider`

## Purpose

Every service built on Raidark exposes the same HTTP and event-flow metrics out of the box, so dashboards and alerts can be standardised at the platform level. The metrics use bounded label cardinality (matched route patterns, not raw URLs) to keep Prometheus stable as the platform scales.

## Configuration

```
METRICS_ENABLED=true        # if false, factory skips registration; the route is also skipped
METRICS_PATH=/metrics       # endpoint path
SERVICE_NAME=my-service     # stamped into logs as the "service" field
```

Wiring (per-service, in `main.go`):

```go
&driverprovider.MetricsProviderFactory{},
```

That factory only registers a `MetricsProvider` on the hub when `METRICS_ENABLED=true`. The `EchoMetricsModule` (registered by Raidark automatically) probes the hub: when no provider is present, the module is a no-op and `/metrics` is not exposed. The `HTTPMetrics` middleware in `EchoApiProvider` does the same probe, so a service without metrics does no per-request work.

## Architecture

```
shared/observability/
├── domain/
│   └── MetricsProvider.go            # interface (Metrics, Handler, Path)
├── driver/
│   └── PrometheusMetricsProvider.go  # concrete impl with private registry
├── metrics.go                        # Metrics bundle (collectors)
└── middleware_metrics.go             # HTTPMetrics(m) Echo middleware

shared/api/driver/modules/
└── EchoMetricsModule.go              # ApiModule that mounts /metrics
```

The interface lives in `domain/`, the implementation in `driver/`, and the route mounting follows Raidark's standard ApiModule pattern. The interface exposes the scrape endpoint as `http.Handler`, not as `(srv *echo.Echo, path string)`, so the contract is framework-neutral and a future non-Echo router could mount it without changes here.

## Metrics

### HTTP

| Name                       | Type      | Labels                  | Description                                          |
|----------------------------|-----------|-------------------------|------------------------------------------------------|
| `http_requests_total`      | counter   | `status`, `endpoint`    | Count of every HTTP response                         |
| `http_request_duration_ms` | histogram | `endpoint`              | Latency in ms, buckets `[5, 25, 100, 500, 1000, 5000]` |

`endpoint` uses Echo's matched route pattern (`/users/:id`), not the raw path (`/users/42`). Routes with no match record `endpoint="unknown"`.

### Domain events

| Name                            | Type      | Labels                              | Description                              |
|---------------------------------|-----------|-------------------------------------|------------------------------------------|
| `events_published_total`        | counter   | `subject`, `outcome`                | Publish attempts                         |
| `events_consumed_total`         | counter   | `subject`, `consumer`, `outcome`    | Consume attempts                         |
| `events_redeliveries_total`     | counter   | `subject`, `consumer`               | Redelivery events from broker            |
| `event_processing_duration_ms`  | histogram | `subject`, `consumer`               | Consumer-side processing latency in ms   |
| `outbox_pending_gauge`          | gauge     | —                                   | Current depth of the transactional outbox |

`outcome` values used across publishers/consumers: `success`, `failure`, `dropped`. Add new ones as the platform evolves; existing labels remain stable.

## Recording metrics

Pull the provider from the hub and call the helpers — they encapsulate label order:

```go
provider := domprovider.Get[obsdomain.MetricsProvider](hub)
metrics := provider.Metrics()

metrics.RecordEventPublished("orders.created", "success")
metrics.RecordEventConsumed("orders.created", "billing-consumer", "success")
metrics.RecordEventRedelivery("orders.created", "billing-consumer")
metrics.ObserveEventProcessing("orders.created", "billing-consumer", elapsedMs)
metrics.SetOutboxPending(currentDepth)
```

Direct access to the underlying `*prometheus.CounterVec` / `*prometheus.HistogramVec` is also available via the `Metrics` struct fields (`HTTPRequestsTotal`, etc.) for advanced cases.

## Why not echo-contrib's prometheus middleware?

The `echo-contrib/echoprometheus` package provides a default HTTP middleware. We considered it and chose to ship our own for three reasons:

1. **Metric names and units are part of the spec.** Raidark mandates `http_requests_total` and `http_request_duration_ms` (milliseconds, not seconds). `echoprometheus` emits `<namespace>_<subsystem>_request_duration_seconds` etc. Renaming via wrappers eats most of the savings.
2. **Default labels include the raw URL.** Without overriding `LabelFuncs`, every distinct URL spawns a unique time series — exactly the cardinality explosion we're trying to avoid. Customising LabelFuncs is non-trivial and equivalent in size to what we wrote.
3. **It doesn't help with event metrics.** Half the spec — `events_published_total`, `events_consumed_total`, `outbox_pending_gauge`, etc. — is independent of the HTTP layer. Adding `echoprometheus` would mean two registries to manage in tandem.

For the HTTP slice alone our middleware is ~30 lines (`middleware_metrics.go`). The maintenance trade-off favours keeping it.

## Testing

Each `MetricsProvider` owns its own `prometheus.Registry`, so unit tests can construct a private provider without colliding with another test's collectors:

```go
m := observability.NewMetrics()
m.RecordEventPublished("test.subject", "success")

count := testutil.ToFloat64(m.EventsPublishedTotal.WithLabelValues("test.subject", "success"))
require.Equal(t, 1.0, count)
```

`prometheus/testutil` (`github.com/prometheus/client_golang/prometheus/testutil`) provides `ToFloat64`, `CollectAndCompare`, and other helpers. Prefer those over reading `/metrics` text in tests.

## Cardinality discipline

Avoid labels that take unbounded user input (raw paths with parameters, free-form error messages, IP addresses). Each unique label-value combination becomes a separate time series; Prometheus scrape size and retention costs scale linearly with cardinality. The `endpoint` label is bounded by the count of registered routes; `subject` by the count of declared event subjects; `consumer` by the count of registered subscribers — all of which grow slowly and predictably.

## Out of scope (for RDK-003)

- OpenTelemetry exporters (OTLP). The W3C trace plumbing is in place so the next step can integrate OTel without touching handlers.
- Grafana dashboards.
- Alerting rules.

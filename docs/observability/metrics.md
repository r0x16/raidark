# Prometheus metrics

Package: `github.com/r0x16/Raidark/shared/observability`

## Purpose

Every service built on Raidark exposes the same HTTP and event-flow metrics out of the box, so dashboards and alerts can be standardised at the platform level. The metrics use bounded label cardinality (matched route patterns, not raw URLs) to keep Prometheus stable as the platform scales.

## Configuration

```
METRICS_ENABLED=true        # if false, /metrics is not mounted; recording calls are still safe
METRICS_PATH=/metrics       # endpoint path
SERVICE_NAME=my-service     # stamped into logs; not yet a metric label
```

The `MetricsProvider` is registered automatically by `MetricsProviderFactory` at boot, ahead of the API provider. `EchoApiProvider`:

1. Pulls the provider from the hub.
2. Registers `observability.HTTPMetrics` after `CorrelationID` and `W3CTrace`.
3. Mounts the scrape endpoint via `provider.MountScrapeEndpoint(server, METRICS_PATH)`.

Services consuming the hub do not need any extra wiring.

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
metrics := domprovider.Get[observability.MetricsProvider](hub).Metrics()

metrics.RecordEventPublished("orders.created", "success")
metrics.RecordEventConsumed("orders.created", "billing-consumer", "success")
metrics.RecordEventRedelivery("orders.created", "billing-consumer")
metrics.ObserveEventProcessing("orders.created", "billing-consumer", elapsedMs)
metrics.SetOutboxPending(currentDepth)
```

Direct access to the underlying `*prometheus.CounterVec` / `*prometheus.HistogramVec` is also available via the `Metrics` struct fields (`HTTPRequestsTotal`, etc.) for advanced cases.

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

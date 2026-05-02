# W3C trace context

Package: `github.com/r0x16/Raidark/shared/observability`

## Purpose

Trace context lets a single request be followed across process boundaries: HTTP gateway → service → message bus → consumer. Raidark uses the [W3C Trace Context](https://www.w3.org/TR/trace-context/) wire format so the same identifiers can be consumed by any compatible backend (Tempo, Jaeger, Honeycomb, Datadog) without rewriting middlewares.

## Headers

| Header             | Direction         | Notes                                                     |
|--------------------|-------------------|-----------------------------------------------------------|
| `traceparent`      | request, response | Mandatory shape: `<version>-<trace-id>-<span-id>-<flags>` |
| `tracestate`       | request, response | Vendor-specific extensions, propagated verbatim           |
| `X-Correlation-ID` | request, response | Canonical Raidark request identifier (RDK-002)            |

Only version `00` is supported. Malformed values are rejected and a fresh trace is generated — per spec, receivers MUST NOT propagate broken values.

## Middleware: `observability.W3CTrace`

Registered automatically by `EchoApiProvider.Setup()` after `rest.CorrelationID()`. It:

1. Reads `traceparent` from the request. If valid, adopts its `trace_id`. The `span_id` it carries is recorded as the parent.
2. If absent, falls back to `X-Correlation-ID` (UUIDv7 with dashes stripped becomes a valid 32-char hex `trace_id`).
3. If the correlation ID can't be promoted, generates a fresh `trace_id` from `crypto/rand`.
4. Always generates a fresh `span_id` for the local span — the entry into this service is a new span regardless of trace inheritance.
5. Stores `trace_id`, `span_id`, `trace_flags`, `trace_state` in both the `echo.Context` (keyed by `observability.ContextTraceIDKey` etc.) and the request's `context.Context` (so `log.FromContext(ctx)` and downstream goroutines can read them).
6. Echoes the resolved `traceparent` (and `tracestate`, if present) on the response, so callers can see what the server is using.

### Framework-agnostic context keys

`ContextTraceIDKey`, `ContextSpanIDKey`, `ContextTraceFlagsKey` and `ContextTraceStateKey` are plain string constants. They are intended to work with any web layer that exposes a get/set bag — not just Echo. Code that runs outside an HTTP server should read the values from the Go context via `GetTraceID(ctx)` etc. instead of going through the framework.

## Reading trace fields

From a Go context (preferred — works everywhere):

```go
traceID := observability.GetTraceID(ctx)
spanID := observability.GetSpanID(ctx)
flags := observability.GetTraceFlags(ctx)
state := observability.GetTraceState(ctx)
```

From `echo.Context` directly (handlers that bypass the request context):

```go
traceID, _ := c.Get(observability.ContextTraceIDKey).(string)
```

## Propagating across transports

`HeaderCarrier` is a two-method interface (`Get`, `Set`). Both `http.Header` and `nats.Header` satisfy it natively — they are both `map[string][]string` aliases with the right method signatures — so no adapter is required for the canonical transports.

For raw `map[string]string` (headers as plain key/value pairs, common in tests and some message bus libraries) use `MapCarrier`:

```go
// Publisher side — NATS
headers := nats.Header{}
observability.InjectTrace(ctx, headers)
natsMsg.Header = headers

// Consumer side — NATS
ctx = observability.ExtractTrace(ctx, msg.Header)

// Plain-map fallback (tests)
headers := observability.MapCarrier{}
observability.InjectTrace(ctx, headers)
```

`InjectTrace` is a no-op when `ctx` has no `trace_id` — it never invents one at injection time, since that would create an orphan trace that can't be joined to the originating request.

This is the helper consumed by RDK-009 / RDK-010 (NATS publisher / consumer drivers) and RDK-016 (HTTP outbound client) to keep `traceparent` flowing across every Raidark transport.

## ID promotion: X-Correlation-ID → trace_id

Raidark already populates `X-Correlation-ID` with a UUIDv7 (RDK-002) — that value is the canonical Raidark request identifier. To bridge legacy callers that don't speak `traceparent` yet, `W3CTrace` strips dashes from the correlation ID and uses it as the `trace_id` when:

- `traceparent` is absent or invalid.
- The correlation ID, after stripping dashes, is exactly 32 lowercase hex characters.
- The result is not all zeros.

UUIDv7 satisfies all three. Any other UUID variant (v4, v1) also qualifies — the conversion only validates shape, not version. This means a UUID-shaped `X-Correlation-ID` set by an upstream gateway flows through as a stable `trace_id` end-to-end without requiring W3C support upstream.

## Frontend / API client integration

The simplest pattern: clients send `X-Correlation-ID: <uuid>` and let `W3CTrace` promote it. No client-side W3C support required.

Clients that already speak `traceparent` (browsers with OpenTelemetry JS, OTel-instrumented services) can send it directly; the existing correlation ID flow continues to work in parallel.

## Why didn't we use OpenTelemetry from day one?

The OTel Go SDK is the long-term destination. We chose not to adopt it in this iteration because:

- **Budget vs. benefit.** OTel brings a substantial dependency footprint (otel-trace, otel-propagation, optionally otel-exporter-otlp-traceparent) plus required configuration scaffolding (tracer providers, span processors, exporters). For "produce a `trace_id` and stamp it on logs," that's expensive.
- **No exporter required yet.** The spec for RDK-003 explicitly leaves OTLP exporters out of scope. Without a backend to export to, the SDK's value is reduced to its propagation helpers — a tiny slice of what it offers.
- **Wire-compatible.** Our types (`TraceContext`, `traceparent` parsing/formatting, `HeaderCarrier`) match the OTel spec exactly. When OTel is introduced, the integration replaces our middleware and helpers without any change to handlers or call-sites that read `observability.GetTraceID(ctx)`.

In short: we built the minimum viable W3C plumbing today, and made it shape-compatible with OTel for tomorrow. The trade-off is the ~150 lines under `shared/observability/{trace,propagation,middleware_trace}.go`.

## Out of scope

- Sampling decisions (currently always `01` / sampled).
- OpenTelemetry SDK integration (OTLP export, span timings, attributes).
- Baggage (`baggage` header) propagation.

These are deliberately deferred to a follow-up so that RDK-003 can land the wire format and the data plumbing without committing to a specific OTel runtime.

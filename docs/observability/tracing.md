# W3C trace context

Package: `github.com/r0x16/Raidark/shared/observability`

## Purpose

Trace context lets a single request be followed across process boundaries: HTTP gateway → service → message bus → consumer. Raidark uses the [W3C Trace Context](https://www.w3.org/TR/trace-context/) wire format so the same identifiers can be consumed by any compatible backend (Tempo, Jaeger, Honeycomb, Datadog) without rewriting middlewares.

The implementation lives in `shared/observability` and is wire-compatible with future OpenTelemetry SDK adoption — the Go context keys, header names, and validation rules match the OTel spec, so adding an OTLP exporter later does not require changes to handlers.

## Headers

| Header        | Direction      | Notes                                                       |
|---------------|----------------|-------------------------------------------------------------|
| `traceparent` | request, response | Mandatory shape: `<version>-<trace-id>-<span-id>-<flags>` |
| `tracestate`  | request, response | Vendor-specific extensions, propagated verbatim            |
| `X-Correlation-ID` | request, response | Legacy fallback, see "ID promotion" below             |

Only version `00` is supported. Malformed values are rejected and a fresh trace is generated — per spec, receivers MUST NOT propagate broken values.

## Middleware: `observability.W3CTrace`

Registered automatically by `EchoApiProvider.Setup()` after `rest.CorrelationID()`. It:

1. Reads `traceparent` from the request. If valid, adopts its `trace_id`. The `span_id` it carries is recorded as the parent.
2. If absent, falls back to `X-Correlation-ID` (UUIDv7 with dashes stripped becomes a valid 32-char hex `trace_id`).
3. If the correlation ID can't be promoted, generates a fresh `trace_id` from `crypto/rand`.
4. Always generates a fresh `span_id` for the local span — the entry into this service is a new span regardless of trace inheritance.
5. Stores `trace_id`, `span_id`, `trace_flags`, `trace_state` in both the `echo.Context` (keyed by `observability.EchoTraceIDKey` etc.) and the request's `context.Context` (so `log.FromContext(ctx)` and downstream goroutines can read them).
6. Echoes the resolved `traceparent` (and `tracestate`, if present) on the response, so callers can see what the server is using.

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
traceID, _ := c.Get(observability.EchoTraceIDKey).(string)
```

## Propagating across transports

For non-HTTP transports (NATS, Kafka, in-process queues) use the carrier helpers:

```go
// Publisher side
headers := observability.MapCarrier{}
observability.InjectTrace(ctx, headers)
natsMsg.Header = nats.Header(headers)

// Consumer side
ctx = observability.ExtractTrace(ctx, observability.MapCarrier(natsMsg.Header))
```

`InjectTrace` is a no-op when `ctx` has no `trace_id` — it never invents one at injection time, since that would create an orphan trace that can't be joined to the originating request.

`HeaderCarrier` is a two-method interface (`Get`, `Set`). Adapters for new transports are trivial; `MapCarrier` is provided for `map[string]string`-shaped headers, which covers NATS and most test cases.

## ID promotion: X-Correlation-ID → trace_id

Raidark already populates `X-Correlation-ID` with a UUIDv7 (RDK-002). To bridge legacy callers that don't speak `traceparent` yet, `W3CTrace` strips dashes from the correlation ID and uses it as the `trace_id` when:

- `traceparent` is absent or invalid.
- The correlation ID, after stripping dashes, is exactly 32 lowercase hex characters.
- The result is not all zeros.

UUIDv7 satisfies all three. Any other UUID variant (v4, v1) also qualifies, since the conversion only validates shape, not version. This means a UUIDv4 `X-Correlation-ID` set by an upstream gateway will flow through as a stable `trace_id` end-to-end.

## Frontend / API client integration

The simplest pattern: clients send `X-Correlation-ID: <uuid>` and let `W3CTrace` promote it. No client-side W3C support required.

Clients that already speak `traceparent` (browsers with OpenTelemetry JS, OTel-instrumented services) can send it directly; the existing correlation ID flow continues to work in parallel.

## Out of scope

- Sampling decisions (currently always `01` / sampled).
- OpenTelemetry SDK integration (OTLP export, span timings, attributes).
- Baggage (`baggage` header) propagation.

These are deliberately deferred to a follow-up so that RDK-003 can land the wire format and the data plumbing without committing to a specific OTel runtime.

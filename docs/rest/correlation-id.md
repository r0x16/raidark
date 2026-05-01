# Correlation ID Middleware

Package: `github.com/r0x16/Raidark/shared/api/rest`

## Purpose

Distributed systems generate many log lines per request, spread across multiple services. Without a shared identifier, correlating a client error to a specific server log entry requires guessing by timestamp, which is unreliable at scale.

The `CorrelationID` middleware gives every request a stable, unique identifier from the moment it enters the system. That ID travels in headers, lives in the Echo context, and surfaces in every `RESTError` response via `trace_id` — so a client can hand an engineer a single string that points directly to the relevant log cluster.

## Wire protocol

- **Request header:** `X-Correlation-ID`
- **Response header:** `X-Correlation-ID` (echoed back)

If the client sends `X-Correlation-ID`, the middleware reuses it verbatim (caller-assigned tracing). If it is absent, the middleware generates a new UUIDv7 and uses that for the lifetime of the request.

## Installation

`CorrelationID` is registered automatically by `EchoApiProvider.Setup()` — no action required in services built on top of Raidark.

```go
// Already done inside EchoApiProvider.Setup(). Do not register again.
e.Use(rest.CorrelationID())
```

It is registered after `middleware.Recover()` and before CORS, so every request — including OPTIONS preflight — carries a trace ID from the earliest point in the middleware chain.

If you use Echo directly (bypassing `EchoApiProvider`), register it manually:

```go
e := echo.New()
e.Use(rest.CorrelationID())
```

## API

### `rest.CorrelationID() echo.MiddlewareFunc`

Returns the middleware function. Reads `X-Correlation-ID` from the request, generates a UUIDv7 if absent, stores the result in `echo.Context`, and writes it back in the response header.

### `rest.GetCorrelationID(c echo.Context) string`

Retrieves the correlation ID stored by the middleware. Returns an empty string if the middleware was not installed for the route.

```go
func MyHandler(c echo.Context) error {
    id := rest.GetCorrelationID(c)
    // pass id to downstream service calls or structured log fields
    ...
}
```

## Integration with error envelope

`rest.RenderError` calls `GetCorrelationID` automatically when `RESTError.TraceID` is empty. Because `EchoApiProvider` installs `CorrelationID` globally, `trace_id` is populated in all error responses without any per-handler work.

## ID format

Generated IDs are UUIDv7 (RFC 9562) via `shared/ids.NewV7()`. They are monotonically increasing within the same millisecond, which preserves log ordering. Caller-supplied IDs are accepted as-is without format validation — this allows upstream gateways to use other ID schemes (e.g. UUID v4, W3C trace-ids).

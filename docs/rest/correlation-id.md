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

Register the middleware globally or per-group on the Echo instance:

```go
import "github.com/r0x16/Raidark/shared/api/rest"

e := echo.New()
e.Use(rest.CorrelationID())
```

It must be installed **before** any middleware or handler that calls `rest.RenderError`, because `RenderError` reads the stored ID to populate `trace_id`.

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

`rest.RenderError` calls `GetCorrelationID` automatically when `RESTError.TraceID` is empty. Installing `CorrelationID` middleware is therefore sufficient to populate `trace_id` in all error responses — no handler changes required.

## ID format

Generated IDs are UUIDv7 (RFC 9562) via `shared/ids.NewV7()`. They are monotonically increasing within the same millisecond, which preserves log ordering. Caller-supplied IDs are accepted as-is without format validation — this allows upstream gateways to use other ID schemes (e.g. UUID v4, W3C trace-ids).

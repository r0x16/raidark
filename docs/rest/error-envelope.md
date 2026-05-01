# REST Error Envelope

Package: `github.com/r0x16/Raidark/shared/api/rest`

## Why a standard error envelope

Without a fixed error shape, every service invents its own format. Clients end up with N parsers, debuggers can't correlate traces across services, and API consumers write fragile `if status == 404 try x else try y` branches. A single envelope eliminates all of that.

## Wire shape

```json
{
  "error": {
    "code": "<domain>.<reason>",
    "message": "Human readable description",
    "details": { "field": "..." },
    "trace_id": "01J..."
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `code` | `string` | yes | Namespaced machine-parseable code, e.g. `forum.topic.not_found`. |
| `message` | `string` | yes | Human-readable description safe to display to end users. |
| `details` | `object` | no | Field-level validation context. Omitted when empty. |
| `trace_id` | `string` | no | Correlation ID (UUIDv7). Populated by `CorrelationID` middleware. |

## Sentinels

| Sentinel | HTTP status | `code` |
|----------|-------------|--------|
| `ErrNotFound` | 404 | `common.not_found` |
| `ErrConflict` | 409 | `common.conflict` |
| `ErrForbidden` | 403 | `common.forbidden` |
| `ErrValidation` | 400 | `common.validation_failed` |
| `ErrTransient` | 503 | `common.transient_failure` |
| `ErrPermanent` | 500 | `common.permanent_failure` |
| _(unknown)_ | 500 | `internal.unexpected` |

Unknown errors always produce `internal.unexpected` with a generic message. The original error is **never** exposed to the caller.

## Two patterns for writing error responses

### Pattern A — Direct render (custom code or details)

Call `rest.RenderError` when you need a domain-specific code or field-level `details` that the sentinel cannot express. The response is written immediately; the handler returns the result of `RenderError`.

```go
import "github.com/r0x16/Raidark/shared/api/rest"

func GetTopic(c echo.Context) error {
    topic, err := svc.FindTopic(c.Param("id"))
    if err != nil {
        return rest.RenderError(c, http.StatusNotFound, &rest.RESTError{
            Code:    "forum.topic.not_found",
            Message: "The requested topic does not exist.",
        })
    }
    return c.JSON(http.StatusOK, topic)
}
```

With field-level validation details:

```go
return rest.RenderError(c, http.StatusBadRequest, &rest.RESTError{
    Code:    "user.invalid_email",
    Message: "The email address is not valid.",
    Details: map[string]any{"field": "email", "value": raw},
})
```

### Pattern B — Return sentinel (standard cases)

Return a sentinel error directly when the standard code and message are sufficient. `EchoErrorHandler` (registered globally by `EchoApiProvider`) intercepts the error and calls `MapError` + `RenderError` automatically. This keeps handlers free of boilerplate.

```go
func GetTopic(c echo.Context) error {
    topic, err := svc.FindTopic(id)
    if err != nil {
        return rest.ErrNotFound  // → 404, "common.not_found"
    }
    return c.JSON(http.StatusOK, topic)
}
```

Sentinels can be wrapped with `fmt.Errorf` — `MapError` uses `errors.Is`, so wrapping is transparent:

```go
return fmt.Errorf("topic %s: %w", id, rest.ErrNotFound)  // still maps to 404
```

## EchoErrorHandler

`rest.EchoErrorHandler` is registered as Echo's global HTTP error handler in `EchoApiProvider.Setup()`. No manual setup is required in individual services that use `EchoApiProvider`.

```go
// Registered automatically — do NOT re-register manually.
e.HTTPErrorHandler = rest.EchoErrorHandler
```

Behaviour:
1. If `c.Response().Committed` is true (Pattern A was used — response already written), the handler does nothing.
2. Otherwise, calls `rest.MapError(err)` to resolve status and `RESTError`, then calls `rest.RenderError`.

This means both patterns produce identical wire output — the difference is only where the error-to-envelope conversion happens.

## `rest.MapError`

Translates a sentinel (or wrapped sentinel) to its HTTP status and `*RESTError`. Returns 500 / `internal.unexpected` for any error that does not match a known sentinel.

```go
status, restErr := rest.MapError(err)
return rest.RenderError(c, status, restErr)
```

## What NOT to do

- Do not expose raw `error.Error()` strings to clients — they may contain internal paths, SQL fragments, or credentials.
- Do not use `echo.NewHTTPError` in Raidark code. It is Echo's native error type; Raidark defines its own error conventions via sentinels and `rest.RenderError`. Using `echo.NewHTTPError` would require bridging two error vocabularies.
- Do not call `c.JSON` directly for error responses — callers expect the `{"error": {...}}` wrapper.
- Do not return a sentinel AND call `rest.RenderError` for the same error — only one path should write the response.

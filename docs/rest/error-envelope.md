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

## API

### `rest.RenderError(c echo.Context, status int, e *RESTError) error`

Serializes `e` as `{"error": {...}}` JSON and writes it with the given status code. If `e.TraceID` is empty, it is populated from the correlation ID stored by the `CorrelationID` middleware.

```go
// In a handler:
import "github.com/r0x16/Raidark/shared/api/rest"

func GetTopic(c echo.Context) error {
    topic, err := svc.FindTopic(c.Param("id"))
    if err != nil {
        status, restErr := rest.MapError(err)
        return rest.RenderError(c, status, restErr)
    }
    return c.JSON(http.StatusOK, topic)
}
```

### `rest.MapError(err error) (int, *RESTError)`

Translates a sentinel (or wrapped sentinel) to its HTTP status and `*RESTError`. Returns 500 / `internal.unexpected` for any error that does not match a known sentinel.

```go
// With custom details (e.g. field validation):
restErr := &rest.RESTError{
    Code:    "user.invalid_email",
    Message: "The email address is not valid.",
    Details: map[string]any{"field": "email", "value": raw},
}
return rest.RenderError(c, http.StatusBadRequest, restErr)
```

### Wrapping sentinels

`MapError` uses `errors.Is`, so wrapped sentinels work correctly:

```go
err := fmt.Errorf("topic %s: %w", id, rest.ErrNotFound)
status, restErr := rest.MapError(err) // → 404, "common.not_found"
```

## What NOT to do

- Do not expose raw `error.Error()` strings to clients — they may contain internal paths, SQL fragments, or credentials.
- Do not invent service-specific HTTP status codes outside the sentinel map. Add a new sentinel instead.
- Do not omit `RenderError` and write `c.JSON` directly for error responses — callers expect the `{"error": {...}}` wrapper.

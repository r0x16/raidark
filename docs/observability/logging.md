# Structured logging

Package: `github.com/r0x16/Raidark/shared/observability/log`

## Purpose

Every Raidark service emits logs in the same shape: one JSON object per line, with the W3C trace fields (`trace_id`, `span_id`), the `service` name, and — when applicable — the `event_id` of the domain event being processed. This lets a single `kubectl logs | jq` (or Loki query) follow a request from the HTTP edge to the consumer that handled the event it triggered, even if the two run in different processes.

The package wraps the existing `domain.LogProvider` contract, so existing call sites that hold a `LogProvider` keep working. The new behaviour — automatic correlation fields — is opt-in via `log.FromContext(ctx)`.

## Selecting the provider

Set the env var:

```
LOGGER_TYPE=observability
LOG_FORMAT=json     # or "text" for local development
LOG_LEVEL=INFO      # DEBUG | INFO | WARNING | ERROR | CRITICAL
SERVICE_NAME=my-service
```

`LOGGER_TYPE=stdout` (the default) keeps the legacy `StdOutLogManager` for services that haven't migrated. Both implementations satisfy `domain.LogProvider`.

`SERVICE_NAME` is registered as a process-wide default. `log.FromContext` will stamp it on every line; if you also call `observability.WithServiceName(ctx, name)`, the per-context value wins.

## Auto-injected fields

`log.FromContext(ctx)` reads the following from `ctx` and adds them as log attributes:

| Attribute   | Source                                     | When emitted                            |
|-------------|--------------------------------------------|------------------------------------------|
| `trace_id`  | `observability.GetTraceID(ctx)`            | Set by `W3CTrace` middleware             |
| `span_id`   | `observability.GetSpanID(ctx)`             | Set by `W3CTrace` middleware             |
| `service`   | `observability.GetServiceName(ctx)` or default | Set by `SERVICE_NAME` or `WithServiceName` |
| `event_id`  | `observability.GetEventID(ctx)`            | Set by event consumer/publisher adapters |

Empty strings are not emitted: `service` defaults to whatever was registered with `SetDefaultServiceName`; absent fields are simply omitted from the JSON object.

## Usage

### Inside an HTTP handler

```go
func MyHandler(c echo.Context, log *obslog.Logger) error {
    log.FromContext(c.Request().Context()).Info("processing", map[string]any{
        "user_id": userID,
    })
    ...
}
```

### Inside an event consumer

```go
func (h *Handler) Handle(ctx context.Context, ev domain.Event) error {
    ctx = observability.WithEventID(ctx, ev.ID)
    h.log.FromContext(ctx).Info("event received", map[string]any{"subject": ev.Subject})
    ...
}
```

### Adding static fields

```go
moduleLog := log.With(map[string]any{"module": "billing"})
moduleLog.FromContext(ctx).Error("charge failed", map[string]any{"order_id": id})
```

`With` returns a new `*Logger`; the receiver is unchanged.

## Format

In JSON mode (`LOG_FORMAT=json`):

```json
{
  "time": "2026-05-01T18:42:11.018-03:00",
  "level": "INFO",
  "source": {"function":"...", "file":"...", "line":42},
  "msg": "processing",
  "trace_id": "0196b3a21c4f7e3da5f20123456789ab",
  "span_id": "5d8e6f10a2b3c4d5",
  "service": "raidark",
  "event_id": "0196b3a2-aaaa-7e3d-a5f2-0123456789ab",
  "user_id": "..."
}
```

In text mode (`LOG_FORMAT=text`) the same fields are emitted as `key=value` pairs. JSON is the production default.

## Log levels

`SetLogLevel(level)` filters subsequent calls. The slog handler is constructed with the boot-time level; at runtime, level changes are enforced in software (matching the legacy `StdOutLogManager` behaviour).

| Level     | Method                              |
|-----------|-------------------------------------|
| `Debug`   | `log.Debug(msg, data)`              |
| `Info`    | `log.Info(msg, data)`               |
| `Warning` | `log.Warning(msg, data)`            |
| `Error`   | `log.Error(msg, data)`              |
| `Critical`| `log.Critical(msg, data)` (→ ERROR) |

## Sensitive data

Unlike `StdOutLogManager`, this logger does not run a `LogDataSanitizer` over `data`. Call sites that need redaction should redact before logging or wrap the logger with their own sanitiser. Future work may unify the two paths once consumers migrate.

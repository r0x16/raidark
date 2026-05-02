# CORS Configuration

Raidark mounts the Echo CORS middleware only when `CORS_ALLOW_ORIGINS` is explicitly set.
If the variable is absent the middleware is **not mounted at all** — preflight requests will
receive no CORS headers and `OPTIONS` responses will fall through to Echo's 404 handler.
This is a deliberate opt-out: services behind a BFF or reverse-proxy that handles CORS
should not carry a second, redundant CORS layer.

## Variables

| Variable | Type | Default | Description |
|---|---|---|---|
| `CORS_ALLOW_ORIGINS` | comma-separated strings | *(not set → CORS disabled)* | Allowed origins. Empty entries are silently dropped. |
| `CORS_ALLOW_HEADERS` | comma-separated strings | `Content-Type,Authorization,X-Requested-With,Accept,Origin` | Headers the browser may include in cross-origin requests. |
| `CORS_ALLOW_METHODS` | comma-separated strings | `GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD` | HTTP methods allowed for cross-origin requests. |
| `CORS_ALLOW_CREDENTIALS` | bool | `false` | Whether the browser may include credentials (cookies, Authorization). |
| `CORS_MAX_AGE` | int (seconds) | `0` (browser default) | Duration for which the preflight response may be cached. `0` defers to the browser's default. |

## Boot log

When the application starts it logs one of:

```
Bootstrap: CORS middleware not mounted   cors=disabled
Bootstrap: CORS middleware configured    cors=https://a.example, https://b.example  ...
```

## Example

```env
CORS_ALLOW_ORIGINS=https://app.example.com,https://admin.example.com
CORS_ALLOW_HEADERS=Content-Type,Authorization,X-Correlation-ID
CORS_ALLOW_METHODS=GET,POST,PUT,PATCH,DELETE,OPTIONS
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=600
```

## Notes

- `CORS_ALLOW_CREDENTIALS=true` with a wildcard origin is rejected by browsers. Always
  pair credentials with an explicit origin list.
- `CORS_MAX_AGE` maps to the `Access-Control-Max-Age` preflight cache header. Setting it
  reduces preflight round-trips in high-traffic browser clients.
- Per-route CORS overrides are **out of scope** for Raidark. Apply them at the reverse-proxy
  layer if needed.

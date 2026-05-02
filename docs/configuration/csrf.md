# CSRF Configuration

Raidark provides built-in CSRF protection via Echo's CSRF middleware. It is **disabled by
default** because services that live behind a BFF (Backend-For-Frontend) which already
enforces CSRF should not add a second, contradictory protection layer.

When `CSRF_ENABLED=false` (default):

- The CSRF middleware is **not mounted** — all requests pass through without token validation.
- The `/csrf-token` route is **not registered** — GET `/csrf-token` returns Echo's 404.

When `CSRF_ENABLED=true`:

- The CSRF middleware validates every non-safe request (POST, PUT, PATCH, DELETE).
- A cookie containing the token is set on the first request.
- The `/csrf-token` endpoint returns the current token for frontend clients that need to
  read it explicitly.

## Variables

| Variable | Type | Default | Description |
|---|---|---|---|
| `CSRF_ENABLED` | bool | `false` | Master toggle. `false` = middleware and endpoint not mounted. |
| `CSRF_TOKEN_LENGTH` | int | `32` | Byte length of the randomly generated CSRF token. |
| `CSRF_COOKIE_NAME` | string | `_csrf` | Name of the cookie that stores the CSRF token. |
| `CSRF_COOKIE_SECURE` | bool | `false` | Set `true` in production to restrict the cookie to HTTPS. |
| `CSRF_TOKEN_LOOKUP` | string | `cookie:_csrf` | Where Echo reads the submitted token. Supports `header:X-CSRF-Token`, `form:csrf`, etc. |
| `CSRF_COOKIE_MAX_AGE` | int (seconds) | `86400` | Lifetime of the CSRF cookie (24 h by default). |

## Boot log

When the application starts it logs one of:

```
Bootstrap: CSRF middleware not mounted   csrf=disabled
Bootstrap: CSRF middleware configured    csrf=enabled  cookie_name=_csrf  ...
```

## Token flow (when enabled)

1. Browser makes any request → server sets `Set-Cookie: _csrf=<token>; HttpOnly; SameSite=Strict`.
2. Browser reads the token via GET `/csrf-token` (returns `{"csrf_token": "<token>"}`).
3. Browser includes the token in subsequent mutating requests via the channel configured in
   `CSRF_TOKEN_LOOKUP` (e.g., `cookie:_csrf`).
4. Middleware validates; mismatch → 403 with the standard Raidark error envelope.

## Notes

- The cookie is always `HttpOnly` and `SameSite=Strict` regardless of other settings.
  These are security invariants, not configuration knobs.
- Double-submit cookie CSRF is **out of scope**. Raidark uses header/cookie token validation
  only.
- Per-route CSRF overrides are **out of scope**. Use the Skipper function in Echo if a
  specific route needs exemption (custom middleware composition).

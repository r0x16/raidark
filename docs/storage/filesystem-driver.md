# Filesystem Storage Driver

The `filesystem` driver stores objects on the local filesystem. It is the default driver and the recommended choice for development and single-node deployments. Cloud migrations (S3, GCS) require only a change to `STORAGE_DRIVER` and the cloud-specific env vars — no application code changes.

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `STORAGE_DRIVER` | `filesystem` | Driver selector |
| `STORAGE_PUBLIC_ROOT` | `/storage/public` | Absolute path to the public object root |
| `STORAGE_PRIVATE_ROOT` | `/storage/private` | Absolute path to the private object root |
| `STORAGE_PUBLIC_BASE_URL` | _(empty)_ | Base URL prepended to public object keys |
| `STORAGE_SIGNING_SECRET` | _(required)_ | Hex-encoded HMAC secret for signed URLs |
| `STORAGE_SIGNED_URL_DEFAULT_TTL` | `600s` | Default TTL for signed URLs (Go duration string) |

`STORAGE_SIGNING_SECRET` is mandatory. A missing or non-hex value causes the server to fail at startup. Generate a suitable secret with:

```sh
openssl rand -hex 32
```

## Directory Layout

Objects are stored at `{root}/{key}`, preserving the full key path:

```
/storage/
├── public/
│   └── users/avatars/2026/05/0196f3a2-6f8c-7d0e-abc1-000000000001.png
└── private/
    └── invoices/pdfs/2026/05/0196f3a2-6f8c-7d0e-abc1-000000000002.pdf
```

The driver creates intermediate directories automatically on `Put`.

## Streaming Writes

`Put` never buffers the full object in memory. It uses `io.TeeReader` to feed the same byte stream simultaneously to the destination file (via `io.Copy`) and an MD5 hasher. Memory usage stays at O(copy-buffer size) — approximately 32 KB — regardless of object size.

The returned `PutResult.ETag` is the hex-encoded MD5 of the written bytes.

## Signed URLs

Private objects are served through an internal Echo handler registered at `GET /_storage/*`. `SignedURL` generates a relative URL of the form:

```
/_storage/{key}?sig={hmac_hex}&expires={unix_timestamp}
```

### Signature Scheme

The HMAC-SHA256 signature is computed over the message:

```
{key}\n{expires_unix_decimal}
```

where `expires_unix_decimal` is a base-10 Unix timestamp (seconds since epoch). The same formula is applied in `FilesystemStorageProvider.SignedURL` (signer) and `FilesystemSignedUrlHandler` (verifier). Any difference between these two — even a single character — makes all URLs invalid.

### Handler Lifecycle

| Condition | HTTP response |
|-----------|---------------|
| `expires` is missing or not an integer | 403 |
| Current time is past `expires` | 403 |
| `sig` is not valid hex | 403 |
| HMAC does not match | 403 |
| Key not found in private root | 404 |
| Valid signature and file exists | 200 |

The 403 cases return a JSON body `{"error": {"code": "storage.url_expired" \| "storage.invalid_signature", ...}}`. HMAC comparison uses `hmac.Equal` for constant-time evaluation, preventing timing-based attacks.

`http.ServeContent` is used to serve the file, which transparently handles `Range` requests, `If-None-Match`, `If-Modified-Since`, and `304 Not Modified` responses.

## Content-Type Detection

Content-Type is inferred from the key's file extension using `mime.TypeByExtension`. If no mapping is found (e.g. no extension), the response defaults to `application/octet-stream`.

## Delete Idempotency

`Delete` checks both roots and returns `nil` if the key is absent. This matches the behavior of cloud object storage APIs and avoids spurious errors in cleanup workflows.

## Path Traversal Protection

Every filesystem path is constructed as:

```go
filepath.Join(root, filepath.FromSlash(key))
```

After joining, the driver verifies the result starts with `root + os.PathSeparator`. A key such as `../../etc/passwd` would be rejected both by `ValidateKey` (which blocks `..` segments) and by this path-prefix check at the driver level.

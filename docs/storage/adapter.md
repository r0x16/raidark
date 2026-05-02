# Storage Adapter

`shared/storage` provides a driver-agnostic abstraction for binary object storage (images, PDFs, documents, etc.). Application code talks only to the `StorageProvider` interface; changing the underlying storage backend is a bootstrap-level concern.

## Interface

```go
type StorageProvider interface {
    Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (PutResult, error)
    Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)
    Delete(ctx context.Context, key string) error
    SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)
    PublicURL(key string) string
    Exists(ctx context.Context, key string) (bool, error)
}
```

### Methods

| Method | Description |
|--------|-------------|
| `Put` | Stream-write an object; returns ETag and byte count |
| `Get` | Open an object for reading; caller must close the `io.ReadCloser` |
| `Delete` | Remove an object; idempotent (missing key is not an error) |
| `SignedURL` | Return a time-limited URL for private-object access |
| `PublicURL` | Return the permanent CDN URL for a public object |
| `Exists` | Check presence without transferring content |

## Visibility

Objects are stored as either **public** or **private** (set in `PutOptions.Visibility`):

| Visibility | Stored under | Access URL |
|------------|--------------|------------|
| `VisibilityPublic` | `STORAGE_PUBLIC_ROOT` | `PublicURL(key)` → permanent CDN URL |
| `VisibilityPrivate` | `STORAGE_PRIVATE_ROOT` | `SignedURL(ctx, key, ttl)` → time-limited URL |

Public objects are served directly from the configured CDN. Private objects are served by the internal handler (filesystem driver) or the driver's native signed-URL mechanism (S3/MinIO/GCS).

## Key Convention

All storage keys must follow the format:

```
{namespace}/{usage}/{year}/{month}/{uuid}.{ext}
```

| Segment | Description | Example |
|---------|-------------|---------|
| `namespace` | Owning service or domain | `users`, `invoices` |
| `usage` | Object category within the namespace | `avatars`, `attachments` |
| `year` | 4-digit year of upload | `2026` |
| `month` | Zero-padded 2-digit month | `05` |
| `uuid` | UUIDv7 generated at upload time | `0196f3a2-...` |
| `.ext` | File extension (optional) | `.png`, `.pdf` |

### Generating and Validating Keys

```go
import domstorage "github.com/r0x16/Raidark/shared/storage/domain"

// Generate a new key
key, err := domstorage.BuildKey("users", "avatars", ".png")

// Validate an existing key
if err := domstorage.ValidateKey(key); err != nil {
    // handle invalid key
}
```

`ValidateKey` rejects: empty keys, absolute paths, path traversal (`../`), and any key that does not match the four-segment format.

## Registration

Storage is optional — services that do not use it pay zero overhead. To enable it, add `StorageProviderFactory` to your providers list **and** `EchoStorageModule` to your modules list (the module mounts the internal signed-URL handler at `GET /_storage/*`):

```go
import (
    driverprovider "github.com/r0x16/Raidark/shared/providers/driver"
    moduleapi "github.com/r0x16/Raidark/shared/api/driver/modules"
)

app := raidark.New([]domprovider.ProviderFactory{
    &driverprovider.ApiProviderFactory{},
    &driverprovider.StorageProviderFactory{},
})

app.Run([]apidomain.ApiModule{
    &moduleapi.EchoStorageModule{EchoModule: app.RootModule("")},
    // ... your other modules
})
```

Services that only use public objects and never call `SignedURL` can omit `EchoStorageModule` — the provider still works, but the internal handler is not mounted.

Retrieve the provider anywhere via the hub:

```go
import domstorage "github.com/r0x16/Raidark/shared/storage/domain"

storage := domprovider.Get[domstorage.StorageProvider](hub)
```

## Usage Example

```go
ctx := context.Background()

// Store a file
key, _ := domstorage.BuildKey("invoices", "pdfs", ".pdf")
result, err := storage.Put(ctx, key, fileReader, domstorage.PutOptions{
    Visibility:  domstorage.VisibilityPrivate,
    ContentType: "application/pdf",
})

// Generate a 10-minute signed URL
signedURL, err := storage.SignedURL(ctx, key, 10*time.Minute)

// Get a permanent public URL
publicURL := storage.PublicURL(key)  // only valid for VisibilityPublic objects

// Check existence
exists, err := storage.Exists(ctx, key)

// Read the file back
rc, info, err := storage.Get(ctx, key)
defer rc.Close()

// Remove the file
err = storage.Delete(ctx, key)
```

## Drivers

| Driver | Status | Description |
|--------|--------|-------------|
| `filesystem` | Available | Local filesystem with HMAC signed URLs; ideal for development and single-node deployments |
| `s3` / `minio` | Planned | S3-compatible object storage |
| `gcs` | Planned | Google Cloud Storage |

See [filesystem-driver.md](filesystem-driver.md) for driver-specific configuration.

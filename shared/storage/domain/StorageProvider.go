// Package domain defines the StorageProvider interface and supporting types for
// binary object storage. It is driver-agnostic: filesystem, S3, MinIO, and GCS
// implementations all satisfy the same interface.
package domain

import (
	"context"
	"io"
	"time"
)

// StorageProvider is the single entry-point for all object storage operations.
// Callers depend only on this interface; the concrete driver is injected at
// bootstrap via StorageProviderFactory.
type StorageProvider interface {
	// Put writes the content of r to the given key. opts controls visibility and
	// content metadata. Writes are streaming — the driver must not buffer the
	// full body in memory.
	Put(ctx context.Context, key string, r io.Reader, opts PutOptions) (PutResult, error)

	// Get returns a read-closer over the object's bytes plus its metadata.
	// The caller is responsible for closing the returned reader.
	Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error)

	// Delete removes the object identified by key. Delete is idempotent:
	// a missing key is not an error.
	Delete(ctx context.Context, key string) error

	// SignedURL returns a time-limited URL that grants read access to a private
	// object without requiring authentication on each request.
	SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error)

	// PublicURL returns the canonical public URL for an object stored with
	// VisibilityPublic. The URL is not time-limited.
	PublicURL(key string) string

	// Exists reports whether the object identified by key is present in storage.
	Exists(ctx context.Context, key string) (bool, error)
}

// Visibility controls whether an object is served publicly or requires a
// signed URL for access.
type Visibility int

const (
	// VisibilityPublic places the object under the public root; PublicURL is valid.
	VisibilityPublic Visibility = iota
	// VisibilityPrivate places the object under the private root; access requires
	// a signed URL obtained from SignedURL.
	VisibilityPrivate
)

// PutOptions carries per-upload metadata.
type PutOptions struct {
	Visibility  Visibility
	ContentType string
	// Size is optional; supply it when known to allow drivers to pre-allocate
	// or set Content-Length on upstream requests.
	Size int64
}

// PutResult is returned after a successful Put.
type PutResult struct {
	Key       string
	SizeBytes int64
	// ETag is the hex-encoded MD5 of the written bytes (filesystem driver) or
	// the driver-native ETag (S3/MinIO). Used for integrity verification.
	ETag string
}

// ObjectInfo carries read-only metadata about a stored object.
type ObjectInfo struct {
	Key         string
	SizeBytes   int64
	ContentType string
	ModifiedAt  time.Time
}

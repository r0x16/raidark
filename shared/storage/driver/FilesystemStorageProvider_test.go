// Package driver_test verifies the filesystem storage driver through its public
// StorageProvider behavior and the signed URL HTTP handler.
package driver_test

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/r0x16/Raidark/shared/api/rest"
	"github.com/r0x16/Raidark/shared/ids"
	storagedomain "github.com/r0x16/Raidark/shared/storage/domain"
	"github.com/r0x16/Raidark/shared/storage/driver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const largeObjectSize = 50 * 1024 * 1024

// TestFilesystemStorageProvider_putGetRoundTripStreamsLargeObject verifies that
// a 50MB upload and download flow stays streaming and does not retain the full
// object in heap memory.
func TestFilesystemStorageProvider_putGetRoundTripStreamsLargeObject(t *testing.T) {
	provider, roots := newFilesystemProvider(t)
	key := newStorageKey(t, "archives", "backup", ".bin")

	before := currentHeapAlloc()
	result, err := provider.Put(context.Background(), key, &repeatingReader{remaining: largeObjectSize}, storagedomain.PutOptions{
		Visibility:  storagedomain.VisibilityPrivate,
		ContentType: "application/octet-stream",
		Size:        largeObjectSize,
	})
	after := currentHeapAlloc()

	require.NoError(t, err)
	assert.Equal(t, key, result.Key)
	assert.Equal(t, int64(largeObjectSize), result.SizeBytes)
	assert.Equal(t, expectedRepeatingReaderMD5(largeObjectSize), result.ETag)
	assert.Less(t, heapGrowth(before, after), uint64(50*1024*1024))

	reader, info, err := provider.Get(context.Background(), key)
	require.NoError(t, err)
	defer reader.Close()

	readBytes, err := io.Copy(io.Discard, reader)
	require.NoError(t, err)
	assert.Equal(t, int64(largeObjectSize), readBytes)
	assert.Equal(t, key, info.Key)
	assert.Equal(t, int64(largeObjectSize), info.SizeBytes)
	assert.FileExists(t, filepath.Join(roots.private, filepath.FromSlash(key)))
}

// TestFilesystemStorageProvider_respectsVisibilityRootsAndPublicURL ensures
// public and private objects land in different roots and public URLs are stable.
func TestFilesystemStorageProvider_respectsVisibilityRootsAndPublicURL(t *testing.T) {
	provider, roots := newFilesystemProvider(t)
	publicKey := newStorageKey(t, "profiles", "avatar", ".txt")
	privateKey := newStorageKey(t, "profiles", "contract", ".txt")

	_, err := provider.Put(context.Background(), publicKey, strings.NewReader("public"), storagedomain.PutOptions{
		Visibility: storagedomain.VisibilityPublic,
	})
	require.NoError(t, err)
	_, err = provider.Put(context.Background(), privateKey, strings.NewReader("private"), storagedomain.PutOptions{
		Visibility: storagedomain.VisibilityPrivate,
	})
	require.NoError(t, err)

	assert.FileExists(t, filepath.Join(roots.public, filepath.FromSlash(publicKey)))
	assert.NoFileExists(t, filepath.Join(roots.private, filepath.FromSlash(publicKey)))
	assert.FileExists(t, filepath.Join(roots.private, filepath.FromSlash(privateKey)))
	assert.NoFileExists(t, filepath.Join(roots.public, filepath.FromSlash(privateKey)))
	assert.Equal(t, "https://cdn.example.test/assets/"+publicKey, provider.PublicURL(publicKey))
}

// TestFilesystemStorageProvider_deleteAndExistsAreIdempotent documents the
// delete contract chosen by RDK-005: missing keys are not errors.
func TestFilesystemStorageProvider_deleteAndExistsAreIdempotent(t *testing.T) {
	provider, _ := newFilesystemProvider(t)
	key := newStorageKey(t, "documents", "pdf", ".pdf")

	exists, err := provider.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)

	_, err = provider.Put(context.Background(), key, strings.NewReader("content"), storagedomain.PutOptions{
		Visibility: storagedomain.VisibilityPrivate,
	})
	require.NoError(t, err)

	exists, err = provider.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.True(t, exists)

	require.NoError(t, provider.Delete(context.Background(), key))
	exists, err = provider.Exists(context.Background(), key)
	require.NoError(t, err)
	assert.False(t, exists)
	assert.NoError(t, provider.Delete(context.Background(), key))
}

// TestFilesystemSignedURLHandler_servesValidURL covers the happy path for the
// internal static handler used by filesystem signed URLs.
func TestFilesystemSignedURLHandler_servesValidURL(t *testing.T) {
	provider, _ := newFilesystemProvider(t)
	key := newStorageKey(t, "private", "image", ".txt")
	_, err := provider.Put(context.Background(), key, strings.NewReader("secret"), storagedomain.PutOptions{
		Visibility: storagedomain.VisibilityPrivate,
	})
	require.NoError(t, err)

	recorder := serveSignedURL(t, provider, mustSignedURL(t, provider, key))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "secret", recorder.Body.String())
	assert.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
}

// TestFilesystemSignedURLHandler_rejectsExpiredURL verifies that TTL validation
// fails before serving private bytes.
func TestFilesystemSignedURLHandler_rejectsExpiredURL(t *testing.T) {
	provider, _ := newFilesystemProvider(t)
	key := newStorageKey(t, "private", "expired", ".txt")
	_, err := provider.Put(context.Background(), key, strings.NewReader("secret"), storagedomain.PutOptions{
		Visibility: storagedomain.VisibilityPrivate,
	})
	require.NoError(t, err)

	signed := mustSignedURL(t, provider, key)
	values := signed.Query()
	values.Set("expires", "1")
	signed.RawQuery = values.Encode()

	recorder := serveSignedURL(t, provider, signed)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "storage.url_expired",
			"message": "The signed URL has expired or is invalid."
		}
	}`, recorder.Body.String())
}

// TestFilesystemSignedURLHandler_rejectsManipulatedHMAC prevents serving files
// when any signed URL query value has been tampered with.
func TestFilesystemSignedURLHandler_rejectsManipulatedHMAC(t *testing.T) {
	provider, _ := newFilesystemProvider(t)
	key := newStorageKey(t, "private", "tampered", ".txt")
	_, err := provider.Put(context.Background(), key, strings.NewReader("secret"), storagedomain.PutOptions{
		Visibility: storagedomain.VisibilityPrivate,
	})
	require.NoError(t, err)

	signed := mustSignedURL(t, provider, key)
	values := signed.Query()
	values.Set("sig", strings.Repeat("0", 64))
	signed.RawQuery = values.Encode()

	recorder := serveSignedURL(t, provider, signed)

	assert.Equal(t, http.StatusForbidden, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "storage.invalid_signature",
			"message": "The signed URL signature is invalid."
		}
	}`, recorder.Body.String())
}

// TestFilesystemSignedURLHandler_returnsNotFoundForMissingKey separates valid
// authorization from object existence: a valid signature still returns 404.
func TestFilesystemSignedURLHandler_returnsNotFoundForMissingKey(t *testing.T) {
	provider, _ := newFilesystemProvider(t)
	key := newStorageKey(t, "private", "missing", ".txt")

	recorder := serveSignedURL(t, provider, mustSignedURL(t, provider, key))

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.JSONEq(t, `{
		"error": {
			"code": "storage.not_found",
			"message": "The requested object was not found."
		}
	}`, recorder.Body.String())
}

// TestNewFilesystemStorageProvider_rejectsInvalidConfiguration covers startup
// failures before a misconfigured service can accept storage traffic.
func TestNewFilesystemStorageProvider_rejectsInvalidConfiguration(t *testing.T) {
	tests := map[string]map[string]string{
		"invalid-secret": {
			"STORAGE_SIGNING_SECRET": "not-hex",
		},
		"empty-secret": {
			"STORAGE_SIGNING_SECRET": "",
		},
		"invalid-ttl": {
			"STORAGE_SIGNING_SECRET":         "736563726574",
			"STORAGE_SIGNED_URL_DEFAULT_TTL": "forever",
		},
	}

	for name, values := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := driver.NewFilesystemStorageProvider(mapEnv(values))
			assert.Error(t, err)
		})
	}
}

type storageRoots struct {
	public  string
	private string
}

// newFilesystemProvider creates a real filesystem-backed provider rooted in the
// test temp directory so visibility checks can assert actual paths.
func newFilesystemProvider(t *testing.T) (*driver.FilesystemStorageProvider, storageRoots) {
	t.Helper()

	base := t.TempDir()
	roots := storageRoots{
		public:  filepath.Join(base, "public"),
		private: filepath.Join(base, "private"),
	}
	provider, err := driver.NewFilesystemStorageProvider(mapEnv{
		"STORAGE_PUBLIC_ROOT":            roots.public,
		"STORAGE_PRIVATE_ROOT":           roots.private,
		"STORAGE_PUBLIC_BASE_URL":        "https://cdn.example.test/assets/",
		"STORAGE_SIGNING_SECRET":         "7365637265742d666f722d7465737473",
		"STORAGE_SIGNED_URL_DEFAULT_TTL": "10m",
	})
	require.NoError(t, err)

	return provider, roots
}

// newStorageKey builds a deterministic-year key while still using the real
// UUIDv7 helper required by the storage key convention.
func newStorageKey(t *testing.T, namespace, usage, ext string) string {
	t.Helper()

	id, err := ids.NewV7()
	require.NoError(t, err)

	key := namespace + "/" + usage + "/2026/05/" + id + ext
	require.NoError(t, storagedomain.ValidateKey(key))
	return key
}

// mustSignedURL parses the relative URL returned by the provider for handler
// tests that need to mutate or replay its query parameters.
func mustSignedURL(t *testing.T, provider *driver.FilesystemStorageProvider, key string) *url.URL {
	t.Helper()

	rawURL, err := provider.SignedURL(context.Background(), key, time.Hour)
	require.NoError(t, err)

	parsed, err := url.Parse(rawURL)
	require.NoError(t, err)
	return parsed
}

// serveSignedURL mounts only the signed URL route so handler tests exercise the
// same Echo routing pattern used by EchoStorageModule.
func serveSignedURL(t *testing.T, provider *driver.FilesystemStorageProvider, signed *url.URL) *httptest.ResponseRecorder {
	t.Helper()

	e := echo.New()
	e.HTTPErrorHandler = rest.EchoErrorHandler
	e.GET("/_storage/*", driver.NewSignedUrlHandler(provider))

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, signed.String(), nil)
	e.ServeHTTP(recorder, request)
	return recorder
}

// currentHeapAlloc forces a GC cycle before reading live heap allocation. The
// test cares about retained memory, not transient allocations inside io.Copy.
func currentHeapAlloc() uint64 {
	runtime.GC()
	var stats runtime.MemStats
	runtime.ReadMemStats(&stats)
	return stats.Alloc
}

// heapGrowth avoids underflow when the second forced GC leaves less live heap
// than the first measurement.
func heapGrowth(before, after uint64) uint64 {
	if after < before {
		return 0
	}
	return after - before
}

// expectedRepeatingReaderMD5 computes the ETag expected from the same streaming
// source without materializing the large object in memory.
func expectedRepeatingReaderMD5(size int64) string {
	hash := md5.New()
	_, _ = io.Copy(hash, &repeatingReader{remaining: size})
	return hex.EncodeToString(hash.Sum(nil))
}

type repeatingReader struct {
	remaining int64
	offset    int64
}

// Read emits deterministic bytes while holding only the caller-provided buffer.
func (r *repeatingReader) Read(p []byte) (int, error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	for i := range p {
		p[i] = byte((r.offset + int64(i)) % 251)
	}
	r.offset += int64(len(p))
	r.remaining -= int64(len(p))
	return len(p), nil
}

type mapEnv map[string]string

// GetString implements EnvProvider with map-backed overrides and defaults.
func (e mapEnv) GetString(key, defaultValue string) string {
	if value, ok := e[key]; ok && value != "" {
		return value
	}
	return defaultValue
}

func (e mapEnv) GetBool(_ string, defaultValue bool) bool { return defaultValue }
func (e mapEnv) GetInt(_ string, defaultValue int) int    { return defaultValue }
func (e mapEnv) GetFloat(_ string, defaultValue float64) float64 {
	return defaultValue
}
func (e mapEnv) GetSlice(_ string, defaultValue []string) []string {
	return defaultValue
}
func (e mapEnv) GetSliceWithSeparator(_ string, _ string, defaultValue []string) []string {
	return defaultValue
}
func (e mapEnv) IsSet(key string) bool {
	value, ok := e[key]
	return ok && value != ""
}
func (e mapEnv) MustGet(key string) string {
	if value, ok := e[key]; ok {
		return value
	}
	return ""
}

var _ io.Reader = (*repeatingReader)(nil)

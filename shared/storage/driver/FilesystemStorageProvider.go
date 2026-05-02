// Package driver provides concrete storage driver implementations.
// The filesystem driver stores objects on the local filesystem using afero's
// BasePathFs, which constrains all operations to a configured root directory
// and prevents path traversal by construction.
package driver

import (
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/afero"

	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domstorage "github.com/r0x16/Raidark/shared/storage/domain"
)

// FilesystemStorageProvider implements StorageProvider using the local filesystem.
//
// Public objects are stored under a BasePathFs rooted at STORAGE_PUBLIC_ROOT;
// private objects under STORAGE_PRIVATE_ROOT. Using BasePathFs means path
// traversal is prevented at the library level — no manual prefix checks needed.
//
// Signed URLs for private objects are relative paths (/_storage/{key}?sig=...&expires=...)
// served by the internal Echo handler registered via EchoStorageModule.
type FilesystemStorageProvider struct {
	publicFs      afero.Fs // BasePathFs rooted at STORAGE_PUBLIC_ROOT
	privateFs     afero.Fs // BasePathFs rooted at STORAGE_PRIVATE_ROOT
	publicBaseURL string
	signingSecret []byte
	defaultTTL    time.Duration
}

// NewFilesystemStorageProvider constructs a FilesystemStorageProvider from
// environment variables. STORAGE_SIGNING_SECRET must be a non-empty hex string.
func NewFilesystemStorageProvider(env domenv.EnvProvider) (*FilesystemStorageProvider, error) {
	secret, err := hex.DecodeString(env.MustGet("STORAGE_SIGNING_SECRET"))
	if err != nil {
		return nil, fmt.Errorf("storage: STORAGE_SIGNING_SECRET is not valid hex: %w", err)
	}
	if len(secret) == 0 {
		return nil, errors.New("storage: STORAGE_SIGNING_SECRET must not be empty")
	}

	ttlStr := env.GetString("STORAGE_SIGNED_URL_DEFAULT_TTL", "600s")
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		return nil, fmt.Errorf("storage: invalid STORAGE_SIGNED_URL_DEFAULT_TTL %q: %w", ttlStr, err)
	}

	base := afero.NewOsFs()
	return &FilesystemStorageProvider{
		publicFs:      afero.NewBasePathFs(base, env.GetString("STORAGE_PUBLIC_ROOT", "/storage/public")),
		privateFs:     afero.NewBasePathFs(base, env.GetString("STORAGE_PRIVATE_ROOT", "/storage/private")),
		publicBaseURL: env.GetString("STORAGE_PUBLIC_BASE_URL", ""),
		signingSecret: secret,
		defaultTTL:    ttl,
	}, nil
}

// Put writes the content of r to the key's path under the appropriate root.
// The write is streaming — io.TeeReader feeds the MD5 hasher while io.Copy
// writes directly to the file, keeping memory usage at O(io.Copy buffer size).
func (p *FilesystemStorageProvider) Put(ctx context.Context, key string, r io.Reader, opts domstorage.PutOptions) (domstorage.PutResult, error) {
	fs := p.fsFor(opts.Visibility)
	fkey := filepath.FromSlash(key)

	if err := fs.MkdirAll(filepath.Dir(fkey), 0755); err != nil {
		return domstorage.PutResult{}, fmt.Errorf("storage: create directories for %q: %w", key, err)
	}

	f, err := fs.Create(fkey)
	if err != nil {
		return domstorage.PutResult{}, fmt.Errorf("storage: create file for %q: %w", key, err)
	}
	defer f.Close()

	hash := md5.New()
	n, err := io.Copy(f, io.TeeReader(r, hash))
	if err != nil {
		return domstorage.PutResult{}, fmt.Errorf("storage: write %q: %w", key, err)
	}

	return domstorage.PutResult{
		Key:       key,
		SizeBytes: n,
		ETag:      hex.EncodeToString(hash.Sum(nil)),
	}, nil
}

// Get opens the object for reading and returns its metadata.
// It probes the public root first, then the private root.
// The caller must close the returned ReadCloser.
func (p *FilesystemStorageProvider) Get(ctx context.Context, key string) (io.ReadCloser, domstorage.ObjectInfo, error) {
	fkey := filepath.FromSlash(key)
	for _, fs := range []afero.Fs{p.publicFs, p.privateFs} {
		f, err := fs.Open(fkey)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, domstorage.ObjectInfo{}, fmt.Errorf("storage: open %q: %w", key, err)
		}

		stat, err := f.Stat()
		if err != nil {
			f.Close()
			return nil, domstorage.ObjectInfo{}, fmt.Errorf("storage: stat %q: %w", key, err)
		}

		ct := mime.TypeByExtension(filepath.Ext(key))
		if ct == "" {
			ct = "application/octet-stream"
		}
		return f, domstorage.ObjectInfo{
			Key:         key,
			SizeBytes:   stat.Size(),
			ContentType: ct,
			ModifiedAt:  stat.ModTime(),
		}, nil
	}
	return nil, domstorage.ObjectInfo{}, fmt.Errorf("storage: key not found: %q", key)
}

// Delete removes the object from whichever root it lives in.
// Delete is idempotent: a missing key returns nil.
func (p *FilesystemStorageProvider) Delete(ctx context.Context, key string) error {
	fkey := filepath.FromSlash(key)
	for _, fs := range []afero.Fs{p.publicFs, p.privateFs} {
		err := fs.Remove(fkey)
		if err == nil {
			return nil
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("storage: delete %q: %w", key, err)
		}
	}
	return nil
}

// SignedURL generates a time-limited URL pointing to the internal handler at
// /_storage/{key}. The URL is relative (path + query only) so it works
// regardless of the host/scheme the service runs behind.
//
// The HMAC message is "{key}\n{expires_unix_decimal}". Both this method and
// the handler in FilesystemSignedUrlHandler.go must use this identical format.
func (p *FilesystemStorageProvider) SignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	if ttl <= 0 {
		ttl = p.defaultTTL
	}
	expiresAt := time.Now().Add(ttl).Unix()

	u := &url.URL{Path: "/_storage/" + key}
	q := url.Values{}
	q.Set("sig", p.computeHMAC(key, expiresAt))
	q.Set("expires", strconv.FormatInt(expiresAt, 10))
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// PublicURL returns the absolute public URL for the given key.
func (p *FilesystemStorageProvider) PublicURL(key string) string {
	return strings.TrimRight(p.publicBaseURL, "/") + "/" + key
}

// Exists reports whether the object is present in either storage root.
func (p *FilesystemStorageProvider) Exists(ctx context.Context, key string) (bool, error) {
	fkey := filepath.FromSlash(key)
	for _, fs := range []afero.Fs{p.publicFs, p.privateFs} {
		_, err := fs.Stat(fkey)
		if err == nil {
			return true, nil
		}
		if !os.IsNotExist(err) {
			return false, fmt.Errorf("storage: stat %q: %w", key, err)
		}
	}
	return false, nil
}

// fsFor returns the afero.Fs for the given visibility.
func (p *FilesystemStorageProvider) fsFor(v domstorage.Visibility) afero.Fs {
	if v == domstorage.VisibilityPrivate {
		return p.privateFs
	}
	return p.publicFs
}

// computeHMAC returns the hex-encoded HMAC-SHA256 of the canonical signed
// message for key and expiresAt. Must remain identical to the verification
// logic in FilesystemSignedUrlHandler.
func (p *FilesystemStorageProvider) computeHMAC(key string, expiresAt int64) string {
	mac := hmac.New(sha256.New, p.signingSecret)
	fmt.Fprintf(mac, "%s\n%d", key, expiresAt)
	return hex.EncodeToString(mac.Sum(nil))
}

package driver

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
	"github.com/r0x16/Raidark/shared/api/rest"
)

// NewSignedUrlHandler returns an Echo handler that validates the HMAC signature
// and TTL embedded in the URL, then serves the private file using http.ServeContent.
//
// The handler is mounted at GET /_storage/* by EchoStorageModule. It only serves
// files from the private root — public objects are served directly via PublicURL
// and never reach this handler.
func NewSignedUrlHandler(provider *FilesystemStorageProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Echo's wildcard param "*" may include a leading slash.
		key := strings.TrimPrefix(c.Param("*"), "/")

		expiresAt, ok := parseAndValidateExpiry(c.QueryParam("expires"))
		if !ok {
			return signedURLForbidden(c, "storage.url_expired", "The signed URL has expired or is invalid.")
		}

		if !verifyHMAC(c.QueryParam("sig"), key, expiresAt, provider.signingSecret) {
			return signedURLForbidden(c, "storage.invalid_signature", "The signed URL signature is invalid.")
		}

		return servePrivateFile(c, provider.privateFs, key)
	}
}

// parseAndValidateExpiry parses the expires query parameter and checks that the
// URL has not yet expired. Returns (timestamp, true) when valid, (0, false) otherwise.
func parseAndValidateExpiry(s string) (int64, bool) {
	ts, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, false
	}
	return ts, time.Now().Unix() <= ts
}

// verifyHMAC performs a constant-time comparison of the provided hex signature
// against the expected HMAC-SHA256 of "{key}\n{expiresAt}".
// Constant-time comparison prevents timing-based signature oracle attacks.
func verifyHMAC(sig, key string, expiresAt int64, secret []byte) bool {
	actual, err := hex.DecodeString(sig)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, secret)
	fmt.Fprintf(mac, "%s\n%d", key, expiresAt)
	return hmac.Equal(actual, mac.Sum(nil))
}

// servePrivateFile opens the file from the private afero.Fs and serves it.
// http.ServeContent handles Range, ETag, If-None-Match, and 304 responses automatically.
func servePrivateFile(c echo.Context, fs afero.Fs, key string) error {
	f, err := fs.Open(filepath.FromSlash(key))
	if err != nil {
		if os.IsNotExist(err) {
			return rest.RenderError(c, http.StatusNotFound, &rest.RESTError{
				Code:    "storage.not_found",
				Message: "The requested object was not found.",
			})
		}
		return err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return err
	}

	ct := mime.TypeByExtension(filepath.Ext(key))
	if ct == "" {
		ct = "application/octet-stream"
	}
	c.Response().Header().Set("Content-Type", ct)
	http.ServeContent(c.Response(), c.Request(), key, stat.ModTime(), f)
	return nil
}

// signedURLForbidden renders a 403 response with a storage-specific RESTError.
func signedURLForbidden(c echo.Context, code, message string) error {
	return rest.RenderError(c, http.StatusForbidden, &rest.RESTError{
		Code:    code,
		Message: message,
	})
}

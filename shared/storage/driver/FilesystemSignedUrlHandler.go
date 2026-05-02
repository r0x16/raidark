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
	"github.com/r0x16/Raidark/shared/api/rest"
)

// NewSignedUrlHandler returns an Echo handler that validates the HMAC signature
// and TTL embedded in the URL, then serves the private file using http.ServeContent.
//
// The handler is mounted at GET /_storage/* by EchoStorageModule. It only serves
// files from the private root — public objects are served directly from the CDN
// via PublicURL and never reach this handler.
func NewSignedUrlHandler(provider *FilesystemStorageProvider) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Echo's wildcard param "*" captures everything after /_storage/
		key := c.Param("*")
		// Trim leading slash that Echo may include in the wildcard capture.
		key = strings.TrimPrefix(key, "/")

		// 1. Parse and validate expiry timestamp.
		expiresStr := c.QueryParam("expires")
		expiresAt, err := strconv.ParseInt(expiresStr, 10, 64)
		if err != nil || time.Now().Unix() > expiresAt {
			return rest.RenderError(c, http.StatusForbidden, &rest.RESTError{
				Code:    "storage.url_expired",
				Message: "The signed URL has expired or is invalid.",
			})
		}

		// 2. Validate HMAC using constant-time comparison to prevent timing attacks.
		sig := c.QueryParam("sig")
		actualBytes, err := hex.DecodeString(sig)
		if err != nil {
			return rest.RenderError(c, http.StatusForbidden, &rest.RESTError{
				Code:    "storage.invalid_signature",
				Message: "The signed URL signature is invalid.",
			})
		}

		mac := hmac.New(sha256.New, provider.signingSecret)
		fmt.Fprintf(mac, "%s\n%d", key, expiresAt)
		expectedBytes := mac.Sum(nil)

		if !hmac.Equal(actualBytes, expectedBytes) {
			return rest.RenderError(c, http.StatusForbidden, &rest.RESTError{
				Code:    "storage.invalid_signature",
				Message: "The signed URL signature is invalid.",
			})
		}

		// 3. Locate and open the file from the private root only.
		fullPath, err := provider.safePath(provider.privateRoot, key)
		if err != nil {
			return rest.RenderError(c, http.StatusForbidden, &rest.RESTError{
				Code:    "storage.invalid_key",
				Message: "The requested key is invalid.",
			})
		}

		f, err := os.Open(fullPath)
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

		// Set Content-Type before ServeContent so it isn't auto-detected from sniffing.
		ct := mime.TypeByExtension(filepath.Ext(key))
		if ct == "" {
			ct = "application/octet-stream"
		}
		c.Response().Header().Set("Content-Type", ct)

		// http.ServeContent handles Range, ETag, If-None-Match, and 304 responses.
		http.ServeContent(c.Response(), c.Request(), key, stat.ModTime(), f)
		return nil
	}
}

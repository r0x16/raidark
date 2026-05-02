package domain

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/r0x16/Raidark/shared/ids"
)

// ValidateKey checks that key conforms to the storage key convention:
//
//	{namespace}/{usage}/{year}/{month}/{uuid}{ext}
//
// It rejects empty keys, absolute paths, path traversal sequences, and any
// key whose structure does not match the four-segment format above.
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("storage: key must not be empty")
	}
	if filepath.IsAbs(key) {
		return fmt.Errorf("storage: key must not be an absolute path: %q", key)
	}
	if strings.Contains(key, "..") {
		return fmt.Errorf("storage: key must not contain path traversal: %q", key)
	}

	// Split on forward slash — the convention always uses forward slashes regardless
	// of the OS path separator.
	parts := strings.SplitN(key, "/", 5)
	if len(parts) != 5 {
		return fmt.Errorf("storage: key %q does not match convention {namespace}/{usage}/{year}/{month}/{uuid}[.ext]", key)
	}

	namespace, usage, yearStr, monthStr, uuidExt := parts[0], parts[1], parts[2], parts[3], parts[4]

	if namespace == "" {
		return fmt.Errorf("storage: key has empty namespace: %q", key)
	}
	if usage == "" {
		return fmt.Errorf("storage: key has empty usage: %q", key)
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil || year < 1970 || len(yearStr) != 4 {
		return fmt.Errorf("storage: key has invalid year %q: %q", yearStr, key)
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil || month < 1 || month > 12 || len(monthStr) != 2 {
		return fmt.Errorf("storage: key has invalid month %q: %q", monthStr, key)
	}

	// Strip optional extension to isolate the UUID stem.
	ext := filepath.Ext(uuidExt)
	uuidStem := strings.TrimSuffix(uuidExt, ext)
	if !ids.IsValidV7(uuidStem) {
		return fmt.Errorf("storage: key UUID segment %q is not a valid UUIDv7: %q", uuidStem, key)
	}

	return nil
}

// BuildKey constructs a canonical storage key for the given namespace, usage,
// and file extension. ext may be empty, or may or may not start with a dot —
// both ".png" and "png" produce the same result.
//
// The returned key follows the convention {namespace}/{usage}/{year}/{month}/{uuid}{ext}.
func BuildKey(namespace, usage, ext string) (string, error) {
	id, err := ids.NewV7()
	if err != nil {
		return "", fmt.Errorf("storage: failed to generate key: %w", err)
	}

	if ext != "" && !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	now := time.Now()
	return fmt.Sprintf(
		"%s/%s/%04d/%02d/%s%s",
		namespace, usage,
		now.Year(), int(now.Month()),
		id, ext,
	), nil
}

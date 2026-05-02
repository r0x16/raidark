package domain

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/r0x16/Raidark/shared/ids"
)

// keyPartCount is the number of slash-separated segments in a valid storage key.
const keyPartCount = 5

// keySegments holds the named components of a parsed storage key.
// This avoids error-prone positional indexing when working with the key parts.
type keySegments struct {
	namespace string
	usage     string
	yearStr   string
	monthStr  string
	uuidExt   string // UUID with optional extension, e.g. "0196f3a2-...uuid.png"
}

// ValidateKey checks that key conforms to the storage key convention:
//
//	{namespace}/{usage}/{year}/{month}/{uuid}[.ext]
//
// It rejects empty keys, absolute paths, path traversal segments ("." or ".."),
// and any key whose structure does not match the four-segment format above.
func ValidateKey(key string) error {
	if key == "" {
		return fmt.Errorf("storage: key must not be empty")
	}
	if filepath.IsAbs(key) {
		return fmt.Errorf("storage: key must not be an absolute path: %q", key)
	}
	// Segment-level traversal check: reject "." and ".." as path components.
	// A name like "a..b" is valid; only isolated dot sequences are dangerous.
	for _, seg := range strings.Split(key, "/") {
		if seg == ".." || seg == "." {
			return fmt.Errorf("storage: key must not contain path traversal segments: %q", key)
		}
	}

	segs, err := parseKeySegments(key)
	if err != nil {
		return err
	}

	if segs.namespace == "" {
		return fmt.Errorf("storage: key has empty namespace: %q", key)
	}
	if segs.usage == "" {
		return fmt.Errorf("storage: key has empty usage: %q", key)
	}

	year, err := strconv.Atoi(segs.yearStr)
	if err != nil || year < 1970 || len(segs.yearStr) != 4 {
		return fmt.Errorf("storage: key has invalid year %q: %q", segs.yearStr, key)
	}

	month, err := strconv.Atoi(segs.monthStr)
	if err != nil || month < 1 || month > 12 || len(segs.monthStr) != 2 {
		return fmt.Errorf("storage: key has invalid month %q: %q", segs.monthStr, key)
	}

	// Strip optional extension to isolate the UUID stem.
	ext := filepath.Ext(segs.uuidExt)
	uuidStem := strings.TrimSuffix(segs.uuidExt, ext)
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

// parseKeySegments splits key on "/" and returns named segments.
// Returns an error if the key does not have exactly keyPartCount parts.
func parseKeySegments(key string) (keySegments, error) {
	parts := strings.SplitN(key, "/", keyPartCount)
	if len(parts) != keyPartCount {
		return keySegments{}, fmt.Errorf(
			"storage: key %q does not match convention {namespace}/{usage}/{year}/{month}/{uuid}[.ext]",
			key,
		)
	}
	return keySegments{
		namespace: parts[0],
		usage:     parts[1],
		yearStr:   parts[2],
		monthStr:  parts[3],
		uuidExt:   parts[4],
	}, nil
}

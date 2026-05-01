package rest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Page is the standard paginated response envelope for list endpoints.
// T is the element type of the result set.
type Page[T any] struct {
	Items      []T      `json:"items"`
	Pagination PageMeta `json:"pagination"`
}

// PageMeta carries the cursor and limit metadata included in every paginated response.
// NextCursor is omitted from JSON when there are no more pages.
type PageMeta struct {
	NextCursor string `json:"next_cursor,omitempty"`
	Limit      int    `json:"limit"`
}

// DefaultLimit is the page size used when the caller omits a limit query parameter.
const DefaultLimit = 20

// MaxLimit is the upper bound enforced by ClampLimit.
const MaxLimit = 100

// EncodeCursor serializes v to an opaque, URL-safe base64 string (no padding).
// v must be JSON-marshallable. The resulting cursor is safe to embed in URLs and
// query strings without further escaping.
func EncodeCursor(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("rest: cursor encode: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

// DecodeCursor reverses EncodeCursor: it base64-decodes the opaque cursor and
// unmarshals the JSON payload into dst. Returns an error if the cursor was tampered
// with or is otherwise malformed.
func DecodeCursor(cursor string, dst any) error {
	data, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return fmt.Errorf("rest: cursor decode: invalid base64: %w", err)
	}
	if err := json.Unmarshal(data, dst); err != nil {
		return fmt.Errorf("rest: cursor decode: invalid payload: %w", err)
	}
	return nil
}

// ClampLimit returns limit unchanged when it is within [1, MaxLimit].
// A zero or negative value is replaced by DefaultLimit; a value above MaxLimit
// is capped at MaxLimit. This prevents clients from requesting unbounded page sizes.
func ClampLimit(limit int) int {
	if limit <= 0 {
		return DefaultLimit
	}
	if limit > MaxLimit {
		return MaxLimit
	}
	return limit
}

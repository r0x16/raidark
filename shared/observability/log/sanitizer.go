package log

import (
	"log/slog"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// DataSanitizer redacts known-sensitive field names and renders complex
// values into compact, length-bounded strings before they are emitted by a
// logger. It is used by both the legacy stdout logger and the
// observability-aware Logger so the redaction policy is enforced uniformly
// regardless of which provider is wired into the hub.
//
// The sanitizer makes two guarantees:
//
//   - Field names containing any fragment in sensitiveKeyFragments (case
//     insensitive) are replaced with "[REDACTED]" — credentials and tokens
//     never reach disk.
//   - Complex values (anything that is not a primitive Go type) are rendered
//     through go-spew with depth 4 and truncated at maxLength characters,
//     so a chatty struct cannot bloat a single log line into megabytes.
type DataSanitizer struct {
	spewConfig *spew.ConfigState
}

// sensitiveKeyFragments is the canonical list of substrings that mark a
// field as sensitive. New fragments should be added here (lowercase) so the
// match is centralized; per-call-site overrides are intentionally not
// supported to keep the policy auditable.
var sensitiveKeyFragments = []string{
	"password",
	"secret",
	"token",
	"authorization",
	"cookie",
	"session_id",
	"certificate",
}

// maxSanitizedLength bounds the rendered representation of complex values.
// Logs are line-oriented in most pipelines; a runaway value can break log
// shippers (Filebeat, Fluent Bit) that have line-size caps.
const maxSanitizedLength = 500

// NewDataSanitizer creates a sanitizer with the canonical spew configuration
// used across Raidark loggers. The configuration favors compact, sorted,
// pointer-stripped output — readable in tooling without dragging in
// allocation addresses or method outputs.
func NewDataSanitizer() *DataSanitizer {
	config := &spew.ConfigState{
		MaxDepth: 4,

		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,

		Indent:           "",
		ContinueOnMethod: false,

		SortKeys: true,
		SpewKeys: false,
	}

	return &DataSanitizer{spewConfig: config}
}

// SanitizeValue passes primitives through verbatim and renders everything
// else via spew, truncating to maxSanitizedLength. The fast path for
// primitives is critical: structured logs are dominated by string and int
// fields, and going through reflection for those would be wasteful.
func (s *DataSanitizer) SanitizeValue(value any) any {
	if value == nil {
		return nil
	}

	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return v
	default:
		safeValue := s.spewConfig.Sprintf("%+v", value)
		if len(safeValue) > maxSanitizedLength {
			return safeValue[:maxSanitizedLength] + "..."
		}
		return safeValue
	}
}

// SanitizeData applies SanitizeField across an entire data map and returns
// a new map; the input is never mutated so callers can safely log the same
// map from multiple goroutines.
func (s *DataSanitizer) SanitizeData(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	sanitized := make(map[string]any, len(data))
	for key, value := range data {
		sanitized[key] = s.SanitizeField(key, value)
	}
	return sanitized
}

// ParseDataForSlog flattens a sanitized data map into the variadic
// []slog.Attr-shaped slice slog.Logger expects. It exists so call sites
// can pass a map[string]any (Raidark's LogProvider contract) directly into
// slog without converting types.
func (s *DataSanitizer) ParseDataForSlog(data map[string]any) []any {
	attrs := make([]any, 0, len(data))
	for key, value := range data {
		sanitizedValue := s.SanitizeField(key, value)
		attrs = append(attrs, slog.Any(key, sanitizedValue))
	}
	return attrs
}

// SanitizeField is the per-field entry point. It checks the key against
// sensitiveKeyFragments first — short-circuiting before any rendering work
// — and otherwise delegates to SanitizeValue. The case fold is done once
// here so SanitizeValue can stay value-only.
func (s *DataSanitizer) SanitizeField(key string, value any) any {
	normalizedKey := strings.ToLower(key)
	for _, fragment := range sensitiveKeyFragments {
		if strings.Contains(normalizedKey, fragment) {
			return "[REDACTED]"
		}
	}
	return s.SanitizeValue(value)
}

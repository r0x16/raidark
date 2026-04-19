package logger

import (
	"log/slog"
	"strings"

	"github.com/davecgh/go-spew/spew"
)

// LogDataSanitizer handles sanitization of complex data structures for logging
type LogDataSanitizer struct {
	spewConfig *spew.ConfigState
}

var sensitiveKeyFragments = []string{
	"password",
	"secret",
	"token",
	"authorization",
	"cookie",
	"session_id",
	"certificate",
}

// NewLogDataSanitizer creates a new instance of LogDataSanitizer with optimal configuration
func NewLogDataSanitizer() *LogDataSanitizer {
	// Custom spew configuration for minimal, readable logging output
	config := &spew.ConfigState{
		// Limit depth to prevent deep nesting
		MaxDepth: 4,

		// Disable verbose information
		DisableMethods:          true,
		DisablePointerMethods:   true,
		DisablePointerAddresses: true,
		DisableCapacities:       true,

		// Compact output format
		Indent:           "",
		ContinueOnMethod: false,

		// Sorting for consistent output
		SortKeys: true,
		SpewKeys: false,
	}

	return &LogDataSanitizer{
		spewConfig: config,
	}
}

// SanitizeValue converts complex values that cannot be JSON serialized to safe representations
func (s *LogDataSanitizer) SanitizeValue(value any) any {
	if value == nil {
		return nil
	}

	// For simple types, return them as-is for performance
	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return v
	default:
		// For complex types, use minimal spew configuration to get basic info only
		safeValue := s.spewConfig.Sprintf("%+v", value)

		// If the output is too long, truncate it
		const maxLength = 500
		if len(safeValue) > maxLength {
			return safeValue[:maxLength] + "..."
		}

		return safeValue
	}
}

// SanitizeData processes a map of data values for safe logging
func (s *LogDataSanitizer) SanitizeData(data map[string]any) map[string]any {
	if data == nil {
		return nil
	}

	sanitized := make(map[string]any, len(data))
	for key, value := range data {
		sanitized[key] = s.sanitizeField(key, value)
	}

	return sanitized
}

// ParseDataForSlog parses data of type map to a slice of slog Attrs with sanitization
func (s *LogDataSanitizer) ParseDataForSlog(data map[string]any) []any {
	attrs := make([]any, 0, len(data))
	for key, value := range data {
		sanitizedValue := s.sanitizeField(key, value)
		attrs = append(attrs, slog.Any(key, sanitizedValue))
	}
	return attrs
}

func (s *LogDataSanitizer) sanitizeField(key string, value any) any {
	normalizedKey := strings.ToLower(key)
	for _, fragment := range sensitiveKeyFragments {
		if strings.Contains(normalizedKey, fragment) {
			return "[REDACTED]"
		}
	}

	return s.SanitizeValue(value)
}

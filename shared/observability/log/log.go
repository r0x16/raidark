// Package log provides a context-aware structured logger that wraps the
// existing shared/logger LogProvider contract. It auto-injects W3C trace
// fields (trace_id, span_id), the service name, and the current event_id
// (when the call site is processing a domain event) into every log line,
// and applies the shared DataSanitizer so sensitive keys are redacted and
// complex values are length-bounded.
//
// This is the recommended logger for every Raidark service. The legacy
// stdout-only logger is still selectable via LOGGER_TYPE=stdout for
// callers that explicitly want no auto-correlation.
package log

import (
	"context"
	"io"
	"log/slog"
	"os"
	"strings"

	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/observability"
)

// Format selects the on-the-wire serialisation of the underlying slog handler.
// It is intentionally not a string alias on env values so callers can't
// accidentally pass arbitrary text where only "json" or "text" is valid.
type Format int

const (
	// FormatJSON emits one JSON object per log line (production default).
	FormatJSON Format = iota
	// FormatText emits the human-readable key=value layout produced by
	// slog.NewTextHandler — useful for local development.
	FormatText
)

// ParseFormat converts the LOG_FORMAT environment string into a Format value.
// Unknown values fall back to JSON: in production we always want machine-
// parseable output, and a typo in an env var should not silently switch us
// back to text.
func ParseFormat(s string) Format {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "text", "txt":
		return FormatText
	default:
		return FormatJSON
	}
}

// Logger is a context-aware LogProvider. It wraps a slog.Logger and pulls
// well-known correlation fields (trace_id, span_id, service, event_id) from
// the context attached at construction time, merging them into the data map
// passed to each call.
//
// Logger satisfies domain.LogProvider so it is a drop-in replacement for the
// legacy StdOutLogManager. The constructor variants (New, NewWithWriter,
// FromContext) cover the production, test and per-request use cases.
type Logger struct {
	logger    *slog.Logger
	level     domlogger.LogLevel
	fields    map[string]any
	sanitizer *DataSanitizer
}

var _ domlogger.LogProvider = (*Logger)(nil)

// New constructs a Logger writing to stdout in the given format and at the
// given log level. Use this at boot time; for per-request loggers prefer
// FromContext, which adds correlation fields without a new handler.
func New(format Format, level domlogger.LogLevel) *Logger {
	return NewWithWriter(os.Stdout, format, level)
}

// NewWithWriter is the same as New but writes to w. Tests use this with a
// bytes.Buffer to assert on the emitted output without touching stdout.
func NewWithWriter(w io.Writer, format Format, level domlogger.LogLevel) *Logger {
	opts := &slog.HandlerOptions{
		Level:     toSlogLevel(level),
		AddSource: true,
	}
	var handler slog.Handler
	switch format {
	case FormatText:
		handler = slog.NewTextHandler(w, opts)
	default:
		handler = slog.NewJSONHandler(w, opts)
	}
	return &Logger{
		logger:    slog.New(handler),
		level:     level,
		sanitizer: NewDataSanitizer(),
	}
}

// FromContext returns a Logger that, in addition to whatever the base logger
// already carries, pulls trace_id, span_id, service and event_id from ctx
// and stamps them on every emitted line. When a field is absent from ctx
// the Logger does not emit the key — empty strings would pollute log
// indexes with hundreds of "" values that mean nothing.
//
// The returned Logger shares the same handler, level and sanitizer as the
// base; only the auto-fields differ. Calling FromContext is therefore cheap
// and safe inside hot paths.
func (l *Logger) FromContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}
	merged := make(map[string]any, len(l.fields)+4)
	for k, v := range l.fields {
		merged[k] = v
	}
	if v := observability.GetTraceID(ctx); v != "" {
		merged["trace_id"] = v
	}
	if v := observability.GetSpanID(ctx); v != "" {
		merged["span_id"] = v
	}
	if v := observability.GetServiceName(ctx); v != "" {
		merged["service"] = v
	}
	if v := observability.GetEventID(ctx); v != "" {
		merged["event_id"] = v
	}
	return &Logger{
		logger:    l.logger,
		level:     l.level,
		fields:    merged,
		sanitizer: l.sanitizer,
	}
}

// With returns a Logger with extra static fields attached. It is the
// equivalent of slog.Logger.With but operates on the data-map shape that
// LogProvider expects so call sites don't have to convert types.
func (l *Logger) With(fields map[string]any) *Logger {
	if len(fields) == 0 {
		return l
	}
	merged := make(map[string]any, len(l.fields)+len(fields))
	for k, v := range l.fields {
		merged[k] = v
	}
	for k, v := range fields {
		merged[k] = v
	}
	return &Logger{
		logger:    l.logger,
		level:     l.level,
		fields:    merged,
		sanitizer: l.sanitizer,
	}
}

// SetLogLevel implements domlogger.LogProvider. The underlying slog handler
// was built with a fixed level (slog handlers don't support dynamic level
// changes without re-construction) so we additionally guard each call site
// with a software-level check, mirroring the StdOutLogManager pattern.
func (l *Logger) SetLogLevel(level domlogger.LogLevel) {
	l.level = level
}

// Debug implements domlogger.LogProvider.
func (l *Logger) Debug(msg string, data map[string]any) {
	if l.level > domlogger.Debug {
		return
	}
	l.logger.Debug(msg, l.attrs(data)...)
}

// Info implements domlogger.LogProvider.
func (l *Logger) Info(msg string, data map[string]any) {
	if l.level > domlogger.Info {
		return
	}
	l.logger.Info(msg, l.attrs(data)...)
}

// Warning implements domlogger.LogProvider.
func (l *Logger) Warning(msg string, data map[string]any) {
	if l.level > domlogger.Warning {
		return
	}
	l.logger.Warn(msg, l.attrs(data)...)
}

// Error implements domlogger.LogProvider.
func (l *Logger) Error(msg string, data map[string]any) {
	l.logger.Error(msg, l.attrs(data)...)
}

// Critical implements domlogger.LogProvider. slog has no Critical level so
// we map it to Error, matching the legacy StdOutLogManager behaviour.
func (l *Logger) Critical(msg string, data map[string]any) {
	l.logger.Error(msg, l.attrs(data)...)
}

// attrs flattens the auto-fields plus the per-call data map into a slice of
// slog attributes, applying the sanitizer to each value. The auto-fields
// are emitted verbatim (they are produced by trusted code) while the data
// map is sanitized to redact sensitive keys and bound complex values.
func (l *Logger) attrs(data map[string]any) []any {
	out := make([]any, 0, 2*(len(l.fields)+len(data)))
	for k, v := range l.fields {
		out = append(out, slog.Any(k, v))
	}
	for k, v := range data {
		out = append(out, slog.Any(k, l.sanitizer.SanitizeField(k, v)))
	}
	return out
}

// toSlogLevel translates the Raidark log-level enum into slog's level scale.
// Critical maps to Error because slog tops out there; the level distinction
// is preserved at the LogProvider level.
func toSlogLevel(level domlogger.LogLevel) slog.Level {
	switch level {
	case domlogger.Debug:
		return slog.LevelDebug
	case domlogger.Info:
		return slog.LevelInfo
	case domlogger.Warning:
		return slog.LevelWarn
	case domlogger.Error, domlogger.Critical:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

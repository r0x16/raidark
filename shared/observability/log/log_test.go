// Package log verifies the observability-aware LogProvider implementation.
package log

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/observability"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoggerFromContext_AddsTraceServiceAndEventFields(t *testing.T) {
	var buffer bytes.Buffer
	base := NewWithWriter(&buffer, FormatJSON, domlogger.Debug).With(map[string]any{
		"component": "worker",
	})
	ctx := context.Background()
	ctx = observability.WithTraceID(ctx, "11111111111111111111111111111111")
	ctx = observability.WithSpanID(ctx, "2222222222222222")
	ctx = observability.WithServiceName(ctx, "billing")
	ctx = observability.WithEventID(ctx, "evt-123")

	base.FromContext(ctx).Info("processed", map[string]any{"attempt": 2})

	entry := decodeLogLine(t, buffer.Bytes())
	assert.Equal(t, "INFO", entry["level"])
	assert.Equal(t, "processed", entry["msg"])
	assert.Equal(t, "11111111111111111111111111111111", entry["trace_id"])
	assert.Equal(t, "2222222222222222", entry["span_id"])
	assert.Equal(t, "billing", entry["service"])
	assert.Equal(t, "evt-123", entry["event_id"])
	assert.Equal(t, "worker", entry["component"])
	assert.EqualValues(t, 2, entry["attempt"])
}

func TestLoggerFromContext_OmitsAbsentAutoFields(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatJSON, domlogger.Info)

	logger.FromContext(context.Background()).Info("processed", nil)

	entry := decodeLogLine(t, buffer.Bytes())
	assert.NotContains(t, entry, "trace_id")
	assert.NotContains(t, entry, "span_id")
	assert.NotContains(t, entry, "event_id")
}

func TestLogger_SanitizesSensitiveData(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatJSON, domlogger.Info)

	logger.Info("login", map[string]any{
		"access_token": "secret-value",
		"safe":         "visible",
	})

	entry := decodeLogLine(t, buffer.Bytes())
	assert.Equal(t, "[REDACTED]", entry["access_token"])
	assert.Equal(t, "visible", entry["safe"])
}

func TestLogger_RespectsConfiguredLevel(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatJSON, domlogger.Error)

	logger.Info("ignored", nil)
	assert.Empty(t, buffer.String())

	logger.Error("emitted", nil)
	entry := decodeLogLine(t, buffer.Bytes())
	assert.Equal(t, "ERROR", entry["level"])
	assert.Equal(t, "emitted", entry["msg"])
}

func TestNew_ConstructsProductionLogger(t *testing.T) {
	assert.NotNil(t, New(FormatJSON, domlogger.Info))
}

func TestNewWithWriter_SupportsTextFormat(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatText, domlogger.Info)

	logger.Info("hello", map[string]any{"component": "test"})

	output := buffer.String()
	assert.Contains(t, output, "level=INFO")
	assert.Contains(t, output, "msg=hello")
	assert.Contains(t, output, "component=test")
}

func TestLoggerFromContext_NilContextReturnsBaseLogger(t *testing.T) {
	logger := NewWithWriter(&bytes.Buffer{}, FormatJSON, domlogger.Info)

	assert.Same(t, logger, logger.FromContext(nil))
}

func TestLogger_WithEmptyFieldsReturnsBaseLogger(t *testing.T) {
	logger := NewWithWriter(&bytes.Buffer{}, FormatJSON, domlogger.Info)

	assert.Same(t, logger, logger.With(nil))
	assert.Same(t, logger, logger.With(map[string]any{}))
}

func TestLogger_LevelGuardsSkipDebugAndWarning(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatJSON, domlogger.Info)

	logger.Debug("ignored debug", nil)
	logger.SetLogLevel(domlogger.Error)
	logger.Warning("ignored warning", nil)

	assert.Empty(t, buffer.String())
}

func TestLogger_WithStaticFieldsAndLevelMethods(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatJSON, domlogger.Debug).With(map[string]any{
		"component": "api",
	}).With(map[string]any{
		"module": "orders",
	})

	logger.Debug("debugging", nil)
	logger.Warning("careful", map[string]any{"step": "warn"})
	logger.Critical("down", nil)

	entries := decodeLogLines(t, buffer.String())
	require.Len(t, entries, 3)
	assert.Equal(t, "DEBUG", entries[0]["level"])
	assert.Equal(t, "api", entries[0]["component"])
	assert.Equal(t, "orders", entries[0]["module"])
	assert.Equal(t, "WARN", entries[1]["level"])
	assert.Equal(t, "warn", entries[1]["step"])
	assert.Equal(t, "ERROR", entries[2]["level"])
	assert.Equal(t, "down", entries[2]["msg"])
}

func TestLogger_SetLogLevelCanRaiseMinimumLevel(t *testing.T) {
	var buffer bytes.Buffer
	logger := NewWithWriter(&buffer, FormatJSON, domlogger.Debug)

	logger.SetLogLevel(domlogger.Warning)
	logger.Info("ignored", nil)
	logger.Warning("visible", nil)

	entries := decodeLogLines(t, buffer.String())
	require.Len(t, entries, 1)
	assert.Equal(t, "WARN", entries[0]["level"])
	assert.Equal(t, "visible", entries[0]["msg"])
}

func TestParseFormat_DefaultsToJSON(t *testing.T) {
	assert.Equal(t, FormatText, ParseFormat(" text "))
	assert.Equal(t, FormatText, ParseFormat("TXT"))
	assert.Equal(t, FormatJSON, ParseFormat(""))
	assert.Equal(t, FormatJSON, ParseFormat("bad-value"))
}

func TestDataSanitizer_SanitizesMapsAttrsAndComplexValues(t *testing.T) {
	sanitizer := NewDataSanitizer()

	assert.Nil(t, sanitizer.SanitizeValue(nil))
	assert.Equal(t, true, sanitizer.SanitizeValue(true))
	assert.Equal(t, map[string]any(nil), sanitizer.SanitizeData(nil))

	sanitized := sanitizer.SanitizeData(map[string]any{
		"Authorization": "Bearer secret",
		"count":         3,
	})
	assert.Equal(t, "[REDACTED]", sanitized["Authorization"])
	assert.Equal(t, 3, sanitized["count"])

	attrs := sanitizer.ParseDataForSlog(map[string]any{"session_id": "abc", "safe": "ok"})
	assert.Len(t, attrs, 2)

	complex := sanitizer.SanitizeValue(map[string]string{"payload": strings.Repeat("x", 700)})
	complexString, ok := complex.(string)
	require.True(t, ok)
	assert.LessOrEqual(t, len(complexString), maxSanitizedLength+3)
	assert.True(t, strings.HasSuffix(complexString, "..."))

	shortComplex := sanitizer.SanitizeValue(struct {
		Name string
	}{Name: "short"})
	shortComplexString, ok := shortComplex.(string)
	require.True(t, ok)
	assert.Contains(t, shortComplexString, `Name:short`)
}

func TestToSlogLevel_MapsEveryRaidarkLevel(t *testing.T) {
	assert.Equal(t, "DEBUG", toSlogLevel(domlogger.Debug).String())
	assert.Equal(t, "INFO", toSlogLevel(domlogger.Info).String())
	assert.Equal(t, "WARN", toSlogLevel(domlogger.Warning).String())
	assert.Equal(t, "ERROR", toSlogLevel(domlogger.Error).String())
	assert.Equal(t, "ERROR", toSlogLevel(domlogger.Critical).String())
	assert.Equal(t, "INFO", toSlogLevel(domlogger.LogLevel(99)).String())
}

func decodeLogLine(t *testing.T, line []byte) map[string]any {
	t.Helper()

	var entry map[string]any
	require.NoError(t, json.Unmarshal(bytes.TrimSpace(line), &entry))
	return entry
}

func decodeLogLines(t *testing.T, output string) []map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(output), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if line == "" {
			continue
		}
		entries = append(entries, decodeLogLine(t, []byte(line)))
	}
	return entries
}

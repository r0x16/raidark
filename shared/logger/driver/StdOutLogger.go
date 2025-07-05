package logger

import (
	"log/slog"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/r0x16/Raidark/shared/logger/domain"
)

// Custom spew configuration for minimal, readable logging output
var minimalSpewConfig = &spew.ConfigState{
	// Limit depth to prevent deep nesting
	MaxDepth: 2,

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

type StdOutLogManager struct {
	logger   *slog.Logger
	logLevel domain.LogLevel
}

var _ domain.LogProvider = &StdOutLogManager{}

func NewStdOutLogManager() *StdOutLogManager {
	manager := &StdOutLogManager{
		logger:   slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		logLevel: domain.Info,
	}
	return manager
}

// Debug implements logger.LogProvider.
func (s *StdOutLogManager) Debug(msg string, data map[string]any) {
	if s.logLevel > domain.Debug {
		return
	}
	s.logger.Debug(msg, s.parseData(data)...)
}

// Info implements logger.LogProvider.
func (s *StdOutLogManager) Info(msg string, data map[string]any) {
	if s.logLevel > domain.Info {
		return
	}
	s.logger.Info(msg, s.parseData(data)...)
}

// Warning implements logger.LogProvider.
func (s *StdOutLogManager) Warning(msg string, data map[string]any) {
	if s.logLevel > domain.Warning {
		return
	}
	s.logger.Warn(msg, s.parseData(data)...)
}

// Error implements logger.LogProvider.
func (s *StdOutLogManager) Error(msg string, data map[string]any) {
	s.logger.Error(msg, s.parseData(data)...)
}

// Critical implements logger.LogProvider.
func (s *StdOutLogManager) Critical(msg string, data map[string]any) {
	s.logger.Error(msg, s.parseData(data)...)
}

// SetLogLevel implements logger.LogProvider.
func (s *StdOutLogManager) SetLogLevel(level domain.LogLevel) {
	s.logLevel = level
}

// Parses data of type map to a slice of slog Attrs.
func (s *StdOutLogManager) parseData(data map[string]any) []any {
	attrs := make([]any, 0, len(data))
	for key, value := range data {
		sanitizedValue := s.sanitizeValue(value)
		attrs = append(attrs, slog.Any(key, sanitizedValue))
	}
	return attrs
}

// sanitizeValue converts complex values that cannot be JSON serialized to safe representations
func (s *StdOutLogManager) sanitizeValue(value any) any {
	if value == nil {
		return nil
	}

	// For simple types, return them as-is for performance
	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return v
	default:
		// For complex types, use minimal spew configuration to get basic info only
		safeValue := minimalSpewConfig.Sprintf("%+v", value)

		// If the output is too long, truncate it
		const maxLength = 500
		if len(safeValue) > maxLength {
			return safeValue[:maxLength] + "..."
		}

		return safeValue
	}
}

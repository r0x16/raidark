package logger

import (
	"log/slog"
	"os"

	"github.com/davecgh/go-spew/spew"
	"github.com/r0x16/Raidark/shared/logger/domain"
)

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

	// Use spew to safely convert complex structures to strings
	// This automatically handles circular references and provides readable output
	safeValue := spew.Sprintf("%+v", value)

	// For simple types, try to return them as-is if possible
	switch v := value.(type) {
	case string, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return v
	default:
		// For complex types, return the safe string representation
		return safeValue
	}
}

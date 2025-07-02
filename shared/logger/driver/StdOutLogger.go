package logger

import (
	"log/slog"
	"os"

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
		attrs = append(attrs, slog.Any(key, value))
	}
	return attrs
}

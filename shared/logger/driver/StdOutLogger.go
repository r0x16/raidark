package logger

import (
	"log/slog"
	"os"

	"github.com/r0x16/Raidark/shared/logger/domain"
)

type StdOutLogManager struct {
	logger    *slog.Logger
	logLevel  domain.LogLevel
	sanitizer *LogDataSanitizer
}

var _ domain.LogProvider = &StdOutLogManager{}

func NewStdOutLogManager(logLevel domain.LogLevel) *StdOutLogManager {
	opts := &slog.HandlerOptions{
		Level:     getSlogLevel(logLevel),
		AddSource: true,
	}
	manager := &StdOutLogManager{
		logger:    slog.New(slog.NewJSONHandler(os.Stdout, opts)),
		logLevel:  logLevel,
		sanitizer: NewLogDataSanitizer(),
	}
	return manager
}

// Debug implements logger.LogProvider.
func (s *StdOutLogManager) Debug(msg string, data map[string]any) {
	if s.logLevel > domain.Debug {
		return
	}
	s.logger.Debug(msg, s.sanitizer.ParseDataForSlog(data)...)
}

// Info implements logger.LogProvider.
func (s *StdOutLogManager) Info(msg string, data map[string]any) {
	if s.logLevel > domain.Info {
		return
	}
	s.logger.Info(msg, s.sanitizer.ParseDataForSlog(data)...)
}

// Warning implements logger.LogProvider.
func (s *StdOutLogManager) Warning(msg string, data map[string]any) {
	if s.logLevel > domain.Warning {
		return
	}
	s.logger.Warn(msg, s.sanitizer.ParseDataForSlog(data)...)
}

// Error implements logger.LogProvider.
func (s *StdOutLogManager) Error(msg string, data map[string]any) {
	s.logger.Error(msg, s.sanitizer.ParseDataForSlog(data)...)
}

// Critical implements logger.LogProvider.
func (s *StdOutLogManager) Critical(msg string, data map[string]any) {
	s.logger.Error(msg, s.sanitizer.ParseDataForSlog(data)...)
}

// SetLogLevel implements logger.LogProvider.
func (s *StdOutLogManager) SetLogLevel(level domain.LogLevel) {
	s.logLevel = level
}

func getSlogLevel(logLevel domain.LogLevel) slog.Level {
	switch logLevel {
	case domain.Debug:
		return slog.LevelDebug
	case domain.Info:
		return slog.LevelInfo
	case domain.Warning:
		return slog.LevelWarn
	case domain.Error:
		return slog.LevelError
	case domain.Critical:
		return slog.LevelError
	}

	return slog.LevelInfo
}

package driver

import (
	"errors"

	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	driverlogger "github.com/r0x16/Raidark/shared/logger/driver"
	obslog "github.com/r0x16/Raidark/shared/observability/log"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type LoggerProviderFactory struct {
	env domenv.EnvProvider
}

func (f *LoggerProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

func (f *LoggerProviderFactory) Register(hub *domain.ProviderHub) error {
	loggerType := f.env.GetString("LOGGER_TYPE", "stdout")
	provider, err := f.getProvider(loggerType)
	if err != nil {
		return err
	}

	domain.Register(hub, provider)
	return nil
}

func (f *LoggerProviderFactory) getProvider(loggerType string) (domlogger.LogProvider, error) {
	level := f.getLogLevel()
	format := obslog.ParseFormat(f.env.GetString("LOG_FORMAT", "json"))
	switch loggerType {
	case "observability":
		// New context-aware logger. Auto-injects trace_id, span_id, service
		// and event_id when callers wrap it with log.FromContext(ctx).
		return obslog.New(format, level), nil
	case "stdout":
		// Legacy logger kept as the default to avoid changing behavior for
		// services that haven't migrated. Selecting "observability" via
		// LOGGER_TYPE opts a service into the trace-aware logger.
		return driverlogger.NewStdOutLogManager(level), nil
	}

	return nil, errors.New("invalid logger type: " + loggerType)
}

func (f *LoggerProviderFactory) getLogLevel() domlogger.LogLevel {
	switch f.env.GetString("LOG_LEVEL", "INFO") {
	case "DEBUG":
		return domlogger.Debug
	case "INFO":
		return domlogger.Info
	case "WARNING":
		return domlogger.Warning
	case "ERROR":
		return domlogger.Error
	case "CRITICAL":
		return domlogger.Critical
	}

	return domlogger.Info
}

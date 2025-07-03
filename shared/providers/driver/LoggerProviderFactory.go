package driver

import (
	"errors"

	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	driverlogger "github.com/r0x16/Raidark/shared/logger/driver"
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
	switch loggerType {
	case "stdout":
		return driverlogger.NewStdOutLogManager(), nil
	}

	return nil, errors.New("invalid logger type: " + loggerType)
}

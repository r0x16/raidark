package driver

import (
	"errors"

	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domevents "github.com/r0x16/Raidark/shared/events/domain"
	driverevents "github.com/r0x16/Raidark/shared/events/driver"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type DomainEventFactory struct {
	envProvider domenv.EnvProvider
	logProvider domlogger.LogProvider
}

var _ domain.ProviderFactory = &DomainEventFactory{}

func (f *DomainEventFactory) Init(hub *domain.ProviderHub) {
	f.envProvider = domain.Get[domenv.EnvProvider](hub)
	f.logProvider = domain.Get[domlogger.LogProvider](hub)
}

func (f *DomainEventFactory) Register(hub *domain.ProviderHub) error {
	domainEventProviderType := f.envProvider.GetString("DOMAIN_EVENT_PROVIDER_TYPE", "in-memory")
	provider, err := f.getProvider(domainEventProviderType, hub)
	if err != nil {
		return err
	}
	domain.Register(hub, provider)
	provider.Collect()
	return nil
}

func (f *DomainEventFactory) getProvider(providerType string, hub *domain.ProviderHub) (domevents.DomainEventsProvider, error) {
	switch providerType {
	case "in-memory":
		bufferSize := f.envProvider.GetInt("DOMAIN_EVENT_BUFFER_SIZE", 100)
		workers := f.envProvider.GetInt("DOMAIN_EVENT_WORKERS", 8)
		provider := driverevents.NewInMemoryDomainEventsProvider(bufferSize, workers, hub)
		return provider, nil
	}

	f.logProvider.Error("invalid domain event provider type", map[string]any{
		"type": providerType,
	})
	return nil, errors.New("invalid domain event provider type: " + providerType)
}

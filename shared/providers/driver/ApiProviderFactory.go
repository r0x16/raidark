package driver

import (
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	driverapi "github.com/r0x16/Raidark/shared/api/driver"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type ApiProviderFactory struct {
	env domenv.EnvProvider
}

func (f *ApiProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

func (f *ApiProviderFactory) Register(hub *domain.ProviderHub) error {
	provider := f.getProvider(hub)
	err := provider.Setup()
	if err != nil {
		return err
	}
	domain.Register(hub, provider)
	return nil
}

func (f *ApiProviderFactory) getProvider(hub *domain.ProviderHub) domapi.ApiProvider {
	return driverapi.NewEchoApiProvider(f.env.GetString("API_PORT", "8080"), hub)
}

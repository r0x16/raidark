package driver

import (
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	driverenv "github.com/r0x16/Raidark/shared/env/driver"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type EnvProviderFactory struct {
}

func (f *EnvProviderFactory) Init(hub *domain.ProviderHub) {
	// EnvProvider doesn't depend on other providers - it's a base dependency
}

func (f *EnvProviderFactory) Register(hub *domain.ProviderHub) error {
	domain.Register(hub, f.getProvider())
	return nil
}

func (f *EnvProviderFactory) getProvider() domenv.EnvProvider {
	return &driverenv.EnvProvider{}
}

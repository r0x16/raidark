package driver

import (
	domauth "github.com/r0x16/Raidark/shared/auth/domain"
	driverauth "github.com/r0x16/Raidark/shared/auth/driver"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type AuthProviderFactory struct {
	env domenv.EnvProvider
}

func (f *AuthProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

func (f *AuthProviderFactory) Register(hub *domain.ProviderHub) error {
	provider := f.getProvider()
	err := provider.Initialize()
	if err != nil {
		return err
	}
	domain.Register(hub, provider)
	return nil
}

func (f *AuthProviderFactory) getProvider() domauth.AuthProvider {
	return driverauth.NewCasdoorAuthProviderFromEnv()
}

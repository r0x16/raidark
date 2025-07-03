package providers

import "github.com/r0x16/Raidark/shared/providers/domain"

type ProviderHubFactory struct {
	hub *domain.ProviderHub
}

func NewProviderHubFactory() *ProviderHubFactory {
	return &ProviderHubFactory{
		hub: &domain.ProviderHub{},
	}
}

func (f *ProviderHubFactory) Create(providers []domain.ProviderFactory) *domain.ProviderHub {
	for _, provider := range providers {
		provider.Init(f.hub)
		provider.Register(f.hub)
	}
	return f.hub
}

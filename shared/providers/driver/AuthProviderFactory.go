package driver

import (
	"fmt"

	domauth "github.com/r0x16/Raidark/shared/auth/domain"
	driverauth "github.com/r0x16/Raidark/shared/auth/driver"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/providers/domain"
)

type AuthProviderFactory struct {
	env domenv.EnvProvider
	log domlogger.LogProvider
}

func (f *AuthProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
	f.log = domain.Get[domlogger.LogProvider](hub)
}

func (f *AuthProviderFactory) Register(hub *domain.ProviderHub) error {
	f.log.Info("Attempting to register AuthProvider", nil)

	authType := f.env.GetString("AUTH_PROVIDER_TYPE", "casdoor")
	f.log.Info("Using AuthProvider type", map[string]any{
		"type": authType,
	})

	provider, err := f.getProvider(authType)
	if err != nil {
		f.log.Error("Failed to create AuthProvider instance", map[string]any{
			"error": err.Error(),
			"type":  authType,
		})
		return fmt.Errorf("failed to create AuthProvider instance: %w", err)
	}

	f.log.Info("Initializing AuthProvider", nil)
	err = provider.Initialize()
	if err != nil {
		f.log.Error("Failed to initialize AuthProvider", map[string]any{
			"error": err.Error(),
		})
		return fmt.Errorf("failed to initialize AuthProvider: %w", err)
	}

	f.log.Info("Successfully initialized AuthProvider, registering in hub", nil)
	domain.Register(hub, provider)

	f.log.Info("AuthProvider successfully registered in hub", nil)
	return nil
}

func (f *AuthProviderFactory) getProvider(authType string) (domauth.AuthProvider, error) {
	switch authType {
	case "casdoor":
		return driverauth.NewCasdoorAuthProviderFromEnv(), nil
	case "array":
		return driverauth.NewArrayAuthProvider(), nil
	default:
		return nil, fmt.Errorf("unsupported auth provider type: %s", authType)
	}
}

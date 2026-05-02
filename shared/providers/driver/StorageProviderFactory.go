package driver

import (
	"fmt"

	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"github.com/r0x16/Raidark/shared/providers/domain"
	domstorage "github.com/r0x16/Raidark/shared/storage/domain"
	storagedriver "github.com/r0x16/Raidark/shared/storage/driver"
)

// StorageProviderFactory registers a StorageProvider in the provider hub.
// The concrete driver is selected by STORAGE_DRIVER (default: "filesystem").
type StorageProviderFactory struct {
	env domenv.EnvProvider
}

// Init implements domain.ProviderFactory.
func (f *StorageProviderFactory) Init(hub *domain.ProviderHub) {
	f.env = domain.Get[domenv.EnvProvider](hub)
}

// Register implements domain.ProviderFactory.
// It registers the StorageProvider interface type so that consumers can retrieve
// the provider without knowing the concrete driver. EchoStorageModule type-asserts
// back to the concrete type only when it needs to mount the internal handler.
func (f *StorageProviderFactory) Register(hub *domain.ProviderHub) error {
	driverName := f.env.GetString("STORAGE_DRIVER", "filesystem")
	switch driverName {
	case "filesystem":
		p, err := storagedriver.NewFilesystemStorageProvider(f.env)
		if err != nil {
			return fmt.Errorf("storage: failed to initialize filesystem driver: %w", err)
		}
		domain.Register[domstorage.StorageProvider](hub, p)
		return nil
	default:
		return fmt.Errorf("storage: unsupported driver %q", driverName)
	}
}

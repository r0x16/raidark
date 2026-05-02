package modules

import (
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	domstorage "github.com/r0x16/Raidark/shared/storage/domain"
	storagedriver "github.com/r0x16/Raidark/shared/storage/driver"
)

// EchoStorageModule registers the internal signed-URL handler for the filesystem
// storage driver. It short-circuits silently when no StorageProvider is in the hub
// (services that don't use storage pay zero overhead) or when the active driver is
// not a FilesystemStorageProvider (cloud drivers sign externally and don't need the
// internal handler).
type EchoStorageModule struct {
	*EchoModule
}

var _ domapi.ApiModule = &EchoStorageModule{}

// Name implements domain.ApiModule.
func (e *EchoStorageModule) Name() string {
	return "Storage"
}

// Setup registers GET /_storage/* when the filesystem driver is active.
func (e *EchoStorageModule) Setup() error {
	if !domprovider.Exists[domstorage.StorageProvider](e.Hub) {
		return nil
	}
	provider := domprovider.Get[domstorage.StorageProvider](e.Hub)
	fsProvider, ok := provider.(*storagedriver.FilesystemStorageProvider)
	if !ok {
		// Non-filesystem drivers (S3, GCS) generate externally signed URLs and
		// do not require this internal handler.
		return nil
	}
	e.Group.GET("/_storage/*", storagedriver.NewSignedUrlHandler(fsProvider))
	return nil
}

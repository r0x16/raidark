package dbmigrate

import (
	"os"

	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/migration/driver/controller"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type Seeder struct {
	modules          []apidomain.ApiModule
	logProvider      domlogger.LogProvider
	databaseProvider domdatastore.DatabaseProvider
}

func NewSeeder(hub *domprovider.ProviderHub, modules []apidomain.ApiModule) *Seeder {
	return &Seeder{
		modules:          modules,
		logProvider:      domprovider.Get[domlogger.LogProvider](hub),
		databaseProvider: domprovider.Get[domdatastore.DatabaseProvider](hub),
	}
}

func (d *Seeder) Run() {
	seedController := &controller.SeederController{
		LogProvider:      d.logProvider,
		DatabaseProvider: d.databaseProvider,
		Modules:          d.modules,
	}
	err := seedController.SeedAction()

	if err != nil {
		// TODO: Catch processing error, and handle it appropriately, such as logging the error or retrying the operation.
		d.logProvider.Critical("Error processing db initialization", map[string]any{"error": err})
		os.Exit(1)
	}
}

package dbmigrate

import (
	"os"

	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	"github.com/r0x16/Raidark/shared/migration/driver/controller"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

type Dbmigrate struct {
	hub         *domprovider.ProviderHub
	modules     []apidomain.ApiModule
	logProvider domlogger.LogProvider
}

func NewDbmigrate(hub *domprovider.ProviderHub, modules []apidomain.ApiModule) *Dbmigrate {
	return &Dbmigrate{
		hub:         hub,
		modules:     modules,
		logProvider: domprovider.Get[domlogger.LogProvider](hub),
	}
}

func (d *Dbmigrate) Run() {

	dbmigrator := &controller.DbMigrationController{
		LogProvider:      d.logProvider,
		DatabaseProvider: domprovider.Get[domdatastore.DatabaseProvider](d.hub),
		Modules:          d.modules,
	}
	err := dbmigrator.MigrateAction()

	if err != nil {
		// TODO: Catch processing error, and handle it appropriately, such as logging the error or retrying the operation.
		d.logProvider.Critical("Error processing db migration", map[string]any{"error": err})
		os.Exit(1)
	}

}

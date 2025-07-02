package dbmigrate

import (
	"fmt"
	"os"

	"github.com/r0x16/Raidark/dbmigrate/controller"
	driverapi "github.com/r0x16/Raidark/shared/api/driver"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	driverdatastore "github.com/r0x16/Raidark/shared/datastore/driver"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	drivelogger "github.com/r0x16/Raidark/shared/logger/driver"
)

type Seeder struct {
}

func NewSeeder() *Seeder {
	return &Seeder{}
}

func (d *Seeder) Run() {
	bundle := &driverapi.ApplicationBundle{
		Database: d.setupDatabase(),
		Log:      d.setupLogger(),
	}
	defer bundle.Database.Close()

	seedController := &controller.SeederController{ApplicationBundle: bundle}
	err := seedController.SeedAction()

	if err != nil {
		// TODO: Catch processing error, and handle it appropriately, such as logging the error or retrying the operation.
		bundle.Log.Critical("Error processing db initialization", map[string]any{"error": err})
		os.Exit(1)
	}

}

func (d *Seeder) setupDatabase() domdatastore.DatabaseProvider {
	dbProvider := &driverdatastore.GormMysqlDatabaseProvider{}
	err := dbProvider.Connect()

	if err != nil {
		fmt.Println("Error connecting to the database:", err)
	}

	return dbProvider
}

func (d *Seeder) setupLogger() domlogger.LogProvider {
	logManager := drivelogger.NewStdOutLogManager()
	level := domlogger.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logManager.SetLogLevel(level)
	return logManager
}

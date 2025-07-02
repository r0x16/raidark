package dbmigrate

import (
	"fmt"
	"os"

	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/dbmigrate/controller"
	"github.com/r0x16/Raidark/shared/domain/logger"
	"github.com/r0x16/Raidark/shared/driver/db"
	stdlog "github.com/r0x16/Raidark/shared/driver/logger"
)

type Seeder struct {
}

func NewSeeder() *Seeder {
	return &Seeder{}
}

func (d *Seeder) Run() {
	bundle := &drivers.ApplicationBundle{
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

func (d *Seeder) setupDatabase() *db.GormMysqlDatabaseProvider {
	dbProvider := &db.GormMysqlDatabaseProvider{}
	err := dbProvider.Connect()

	if err != nil {
		fmt.Println("Error connecting to the database:", err)
	}

	return dbProvider
}

func (d *Seeder) setupLogger() logger.LogProvider {
	logManager := stdlog.NewStdOutLogManager()
	level := logger.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logManager.SetLogLevel(level)
	return logManager
}

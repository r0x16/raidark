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

type Dbmigrate struct{}

func NewDbmigrate() *Dbmigrate {
	return &Dbmigrate{}
}

func (d *Dbmigrate) Run() {
	bundle := &drivers.ApplicationBundle{
		Database: d.setupDatabase(),
		Log:      d.setupLogger(),
	}
	defer bundle.Database.Close()

	dbmigrator := &controller.DbMigrationController{ApplicationBundle: bundle}
	err := dbmigrator.MigrateAction()

	if err != nil {
		// TODO: Catch processing error, and handle it appropriately, such as logging the error or retrying the operation.
		bundle.Log.Critical("Error processing db migration", map[string]any{"error": err})
		os.Exit(1)
	}

}

/*
 * Setup the database connection
 * This method creates a new postgres database provider and connects to the database
 */
func (d *Dbmigrate) setupDatabase() *db.GormPostgresDatabaseProvider {
	dbProvider := &db.GormPostgresDatabaseProvider{}
	err := dbProvider.Connect()

	if err != nil {
		fmt.Println(err)
		panic("Error connecting to the database:")
	}

	return dbProvider
}

/*
 * Setup the logger
 * This method creates a new std out log manager and sets the log level
 */
func (d *Dbmigrate) setupLogger() logger.LogProvider {
	logManager := stdlog.NewStdOutLogManager()
	level := logger.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logManager.SetLogLevel(level)
	return logManager
}

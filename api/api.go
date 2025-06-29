package api

import (
	"fmt"
	"os"

	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/api/drivers/modules"
	"github.com/r0x16/Raidark/api/services"
	"github.com/r0x16/Raidark/shared/domain/logger"
	"github.com/r0x16/Raidark/shared/driver/db"
	"github.com/r0x16/Raidark/shared/driver/events"
)

// TODO: Refactorize

type Api struct{}

func NewApi() *Api {
	return &Api{}
}

func (a *Api) Run() {
	bundle := &drivers.ApplicationBundle{
		Database: a.setupDatabase(),
		Log:      a.setupLogger(),
	}
	defer bundle.Database.Close()

	port := os.Getenv("API_PORT")
	server := drivers.NewEchoApiProvider(port, bundle)

	a.registerModules(server)

	service := services.NewApiService(server, bundle.Log)
	service.Run()

}

func (a *Api) registerModules(server *drivers.EchoApiProvider) {
	server.Register(&modules.EchoMainModule{Api: server})
}

func (d *Api) setupDatabase() *db.GormMysqlDatabaseProvider {
	dbProvider := &db.GormMysqlDatabaseProvider{}
	err := dbProvider.Connect()

	if err != nil {
		fmt.Println(err)
		panic("Error connecting to the database:")
	}

	return dbProvider
}

func (d *Api) setupLogger() *events.StdOutLogManager {
	logManager := events.NewStdOutLogManager()
	level := logger.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logManager.SetLogLevel(level)
	return logManager
}

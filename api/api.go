package api

import (
	"fmt"
	"os"

	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/api/drivers/modules"
	"github.com/r0x16/Raidark/api/services"
	domauth "github.com/r0x16/Raidark/shared/domain/auth"
	"github.com/r0x16/Raidark/shared/domain/logger"
	"github.com/r0x16/Raidark/shared/driver/auth"
	"github.com/r0x16/Raidark/shared/driver/db"
	"github.com/r0x16/Raidark/shared/driver/env"
	stdlog "github.com/r0x16/Raidark/shared/driver/logger"
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
		Auth:     a.setupAuth(),
		Env:      a.setupEnv(),
	}
	defer bundle.Database.Close()

	port := os.Getenv("API_PORT")
	server := drivers.NewEchoApiProvider(port, bundle)

	a.registerModules(server)

	service := services.NewApiService(server, bundle.Log)
	service.Run()

}

/*
 * Register the modules
 * This method registers the modules to the server
 */
func (a *Api) registerModules(server *drivers.EchoApiProvider) {
	server.Register(&modules.EchoMainModule{Api: server})
	server.Register(&modules.EchoAuthModule{Api: server})
	// Add more modules here
}

/*
 * Setup the database connection
 * This method creates a new postgres database provider and connects to the database
 */
func (d *Api) setupDatabase() *db.GormPostgresDatabaseProvider {
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
func (d *Api) setupLogger() logger.LogProvider {
	logManager := stdlog.NewStdOutLogManager()
	level := logger.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logManager.SetLogLevel(level)
	return logManager
}

/*
 * Setup the auth provider
 * This method creates a new casdoor auth provider and connects to the auth provider
 */
func (d *Api) setupAuth() domauth.AuthProvider {
	authProvider := auth.NewCasdoorAuthProviderFromEnv()
	err := authProvider.Initialize()
	if err != nil {
		fmt.Println(err)
		panic("Error initializing the auth provider:")
	}
	return authProvider
}

/*
 * Setup the environment provider
 * This method creates a new environment provider for configuration management
 */
func (d *Api) setupEnv() *env.EnvProvider {
	return env.NewEnvProvider()
}

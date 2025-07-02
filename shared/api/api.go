package api

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	driverapi "github.com/r0x16/Raidark/shared/api/driver"
	apimodules "github.com/r0x16/Raidark/shared/api/driver/modules"
	apiservices "github.com/r0x16/Raidark/shared/api/service"
	"github.com/r0x16/Raidark/shared/auth/domain"
	"github.com/r0x16/Raidark/shared/auth/driver"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	driverdatastore "github.com/r0x16/Raidark/shared/datastore/driver"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	driverenv "github.com/r0x16/Raidark/shared/env/driver"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	drivelogger "github.com/r0x16/Raidark/shared/logger/driver"
)

// TODO: Refactorize

type Api struct{}

func NewApi() *Api {
	return &Api{}
}

func (a *Api) Run() {
	bundle := &driverapi.ApplicationBundle{
		Database: a.setupDatabase(),
		Log:      a.setupLogger(),
		Auth:     a.setupAuth(),
		Env:      a.setupEnv(),
	}
	defer bundle.Database.Close()

	port := os.Getenv("API_PORT")
	server := driverapi.NewEchoApiProvider(port, bundle)

	a.registerModules(server)

	service := apiservices.NewApiService(server, bundle.Log)
	service.Run()

}

/*
 * Register the modules
 * This method registers the modules to the server
 */
func (a *Api) registerModules(server *driverapi.EchoApiProvider) {

	rootModule := apimodules.EchoModule{
		Api:   server,
		Group: server.Server.Group(""),
	}

	apiv1Module := apimodules.EchoModule{
		Api:   server,
		Group: server.Server.Group("/api/v1"),
	}

	apiv1Module.Group.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:" + echo.HeaderAuthorization,
		AuthScheme: "Bearer",
		Validator: func(key string, c echo.Context) (bool, error) {
			token, err := server.Bundle.Auth.ParseToken(key)
			if err != nil {
				fmt.Println(err)
				return false, err
			}
			c.Set("user", token)
			return true, nil
		},
	}))

	server.Register(&apimodules.EchoMainModule{EchoModule: rootModule})
	server.Register(&apimodules.EchoAuthModule{EchoModule: rootModule})
	server.Register(&apimodules.EchoApiMainModule{EchoModule: apiv1Module})
	// Add more modules here
}

/*
 * Setup the database connection
 * This method creates a new postgres database provider and connects to the database
 */
func (d *Api) setupDatabase() domdatastore.DatabaseProvider {
	dbProvider := &driverdatastore.GormPostgresDatabaseProvider{}
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
func (d *Api) setupLogger() domlogger.LogProvider {
	logManager := drivelogger.NewStdOutLogManager()
	level := domlogger.ParseLogLevel(os.Getenv("LOG_LEVEL"))
	logManager.SetLogLevel(level)
	return logManager
}

/*
 * Setup the auth provider
 * This method creates a new casdoor auth provider and connects to the auth provider
 */
func (d *Api) setupAuth() domain.AuthProvider {
	authProvider := driver.NewCasdoorAuthProviderFromEnv()
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
func (d *Api) setupEnv() domenv.EnvProvider {
	return driverenv.NewEnvProvider()
}

package api

import (
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	"github.com/r0x16/Raidark/shared/api/driver"
	apimodules "github.com/r0x16/Raidark/shared/api/driver/modules"
	apiservices "github.com/r0x16/Raidark/shared/api/service"
	domauth "github.com/r0x16/Raidark/shared/auth/domain"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	providers "github.com/r0x16/Raidark/shared/providers/services"
)

// TODO: Refactorize

type Api struct {
	AuthProvider      domauth.AuthProvider
	LogProvider       domlogger.LogProvider
	DatastoreProvider domdatastore.DatabaseProvider
	EnvProvider       domenv.EnvProvider
	ApiProvider       domapi.ApiProvider
	Hub               *domprovider.ProviderHub
}

func NewApi(ctx context.Context) *Api {
	api := &Api{}
	api.Hub = api.setupHub(ctx)
	api.setupProviders(api.Hub)
	return api
}

func (a *Api) Run() {
	defer a.DatastoreProvider.Close()
	server := a.ApiProvider

	a.registerModules(server)

	service := apiservices.NewApiService(server, a.LogProvider)
	service.Run()

}

func (a *Api) setupProviders(hub *domprovider.ProviderHub) {
	a.LogProvider = domprovider.Get[domlogger.LogProvider](hub)
	a.AuthProvider = domprovider.Get[domauth.AuthProvider](hub)
	a.DatastoreProvider = domprovider.Get[domdatastore.DatabaseProvider](hub)
	a.EnvProvider = domprovider.Get[domenv.EnvProvider](hub)
	a.ApiProvider = domprovider.Get[domapi.ApiProvider](hub)
}

func (a *Api) setupHub(ctx context.Context) *domprovider.ProviderHub {
	hubFactory := providers.NewProviderHubFactory()
	providers := ctx.Value("providers").([]domprovider.ProviderFactory)
	hub := hubFactory.Create(providers)
	return hub
}

/*
 * Register the modules
 * This method registers the modules to the server
 */
func (a *Api) registerModules(server domapi.ApiProvider) {

	echoServer := a.ApiProvider.(*driver.EchoApiProvider)

	rootModule := apimodules.EchoModule{
		Api:   server,
		Group: echoServer.Server.Group(""),
		Hub:   a.Hub,
	}

	apiv1Module := apimodules.EchoModule{
		Api:   server,
		Group: echoServer.Server.Group("/api/v1"),
		Hub:   a.Hub,
	}

	apiv1Module.Group.Use(middleware.KeyAuthWithConfig(middleware.KeyAuthConfig{
		KeyLookup:  "header:" + echo.HeaderAuthorization,
		AuthScheme: "Bearer",
		Validator: func(key string, c echo.Context) (bool, error) {
			token, err := a.AuthProvider.ParseToken(key)
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

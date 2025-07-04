package api

import (
	domapi "github.com/r0x16/Raidark/shared/api/domain"
	apiservices "github.com/r0x16/Raidark/shared/api/service"
	domlogger "github.com/r0x16/Raidark/shared/logger/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
)

// TODO: Refactorize

type Api struct {
	Hub         *domprovider.ProviderHub
	Modules     []domapi.ApiModule
	ApiProvider domapi.ApiProvider
	LogProvider domlogger.LogProvider
}

func NewApi(hub *domprovider.ProviderHub, modules []domapi.ApiModule) *Api {
	api := &Api{}
	api.Hub = hub
	api.Modules = modules
	api.setupProviders()
	return api
}

func (a *Api) Run() {

	a.registerModules(a.ApiProvider, a.Modules)

	service := apiservices.NewApiService(a.ApiProvider, a.LogProvider)
	service.Run()

}

func (a *Api) setupProviders() {
	a.ApiProvider = domprovider.Get[domapi.ApiProvider](a.Hub)
	a.LogProvider = domprovider.Get[domlogger.LogProvider](a.Hub)
}

/*
 * Register the modules
 * This method registers the modules to the server
 */
func (a *Api) registerModules(server domapi.ApiProvider, modules []domapi.ApiModule) {

	for _, module := range modules {
		server.Register(module)
	}
	// Add more modules here
}

package raidark

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	moduleapi "github.com/r0x16/Raidark/shared/api/driver/modules"
	"github.com/r0x16/Raidark/shared/cmd"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domevents "github.com/r0x16/Raidark/shared/events/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	driverprovider "github.com/r0x16/Raidark/shared/providers/driver"
	svcproviders "github.com/r0x16/Raidark/shared/providers/services"
)

// Raidark is the main struct that contains the providers, modules.
// It is used to initialize the providers, modules and run the application.
type Raidark struct {
	providers []domprovider.ProviderFactory
	modules   []apidomain.ApiModule
	hub       *domprovider.ProviderHub
	datastore domdatastore.DatabaseProvider
	events    domevents.DomainEventsProvider
}

// New creates a new Raidark instance.
func New(providers []domprovider.ProviderFactory) *Raidark {
	raidark := &Raidark{
		providers: providers,
	}
	raidark.loadEnvIfExists()
	raidark.hub = raidark.initializeProviders(raidark.providers)
	// Initialize datastore from provider hub
	if domprovider.Exists[domdatastore.DatabaseProvider](raidark.hub) {
		raidark.datastore = domprovider.Get[domdatastore.DatabaseProvider](raidark.hub)
	}

	if domprovider.Exists[domevents.DomainEventsProvider](raidark.hub) {
		raidark.events = domprovider.Get[domevents.DomainEventsProvider](raidark.hub)
	}
	return raidark
}

// registerModules registers the modules to the server
// It adds the root module and the modules to the server
func (r *Raidark) registerModules(modules []apidomain.ApiModule) {
	rootModule := r.RootModule("")
	r.modules = append(r.modules, &moduleapi.EchoMainModule{EchoModule: rootModule})
	r.modules = append(r.modules, modules...)
}

// initializeProviders initializes the providers
// It adds the base providers and the providers to the hub
func (r *Raidark) initializeProviders(providers []domprovider.ProviderFactory) *domprovider.ProviderHub {
	// Add base providers first - they're needed by other providers
	baseProviders := []domprovider.ProviderFactory{
		&driverprovider.EnvProviderFactory{},
		&driverprovider.LoggerProviderFactory{},
	}
	allProviders := append(baseProviders, providers...)

	hubFactory := svcproviders.NewProviderHubFactory()
	hub := hubFactory.Create(allProviders)
	return hub
}

// initializeEventListeners initializes the event listeners
// It adds the event listeners to the event provider
func (r *Raidark) initializeEventListeners(modules []apidomain.ApiModule) {
	for _, module := range modules {
		listeners := module.GetEventListeners()
		for _, listener := range listeners {
			r.events.Subscribe(listener)
		}
	}
}

// Run runs the application
// It registers the modules, initializes the event listeners and executes the command
func (r *Raidark) Run(modules []apidomain.ApiModule) {
	if r.datastore != nil {
		defer r.datastore.Close()
	}
	r.registerModules(modules)
	r.initializeEventListeners(r.modules)
	cmd.Execute(r.hub, r.modules)
}

// RootModule creates a new EchoModule
// It is used to create the root module
func (r *Raidark) RootModule(groupPath string) *moduleapi.EchoModule {
	return moduleapi.NewEchoModule(groupPath, r.hub)
}

// AuthenticatedRootModule creates a new EchoModule
// It is used to create the authenticated root module
func (r *Raidark) AuthenticatedRootModule(groupPath string) *moduleapi.EchoModule {
	return moduleapi.NewAuthenticatedEchoModule(groupPath, r.hub)
}

// loadEnvIfExists loads the environment variables from the .env file
// It is used to load the environment variables from the .env file
func (r *Raidark) loadEnvIfExists() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
		log.Println("Environment file loaded successfully")
	}
}

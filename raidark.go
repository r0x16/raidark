package raidark

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	moduleapi "github.com/r0x16/Raidark/shared/api/driver/modules"
	"github.com/r0x16/Raidark/shared/cmd"
	domdatastore "github.com/r0x16/Raidark/shared/datastore/domain"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	driverprovider "github.com/r0x16/Raidark/shared/providers/driver"
	svcproviders "github.com/r0x16/Raidark/shared/providers/services"
)

type Raidark struct {
	providers []domprovider.ProviderFactory
	modules   []apidomain.ApiModule
	hub       *domprovider.ProviderHub
	datastore domdatastore.DatabaseProvider
}

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
	return raidark
}

func (r *Raidark) registerModules(modules []apidomain.ApiModule) {
	rootModule := r.RootModule("")
	r.modules = append(r.modules, &moduleapi.EchoMainModule{EchoModule: rootModule})
	r.modules = append(r.modules, modules...)
}

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

func (r *Raidark) Run(modules []apidomain.ApiModule) {
	if r.datastore != nil {
		defer r.datastore.Close()
	}
	r.registerModules(modules)
	cmd.Execute(r.hub, r.modules)
}

func (r *Raidark) RootModule(groupPath string) *moduleapi.EchoModule {
	return moduleapi.NewEchoModule(groupPath, r.hub)
}

func (r *Raidark) AuthenticatedRootModule(groupPath string) *moduleapi.EchoModule {
	return moduleapi.NewAuthenticatedEchoModule(groupPath, r.hub)
}

func (r *Raidark) loadEnvIfExists() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Fatalf("Error loading .env file: %v", err)
		}
		log.Println("Environment file loaded successfully")
	}
}

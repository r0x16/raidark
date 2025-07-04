/*
Copyright © 2024 r0x16
*/
package main

import (
	raidark "github.com/r0x16/Raidark"
	apidomain "github.com/r0x16/Raidark/shared/api/domain"
	moduleapi "github.com/r0x16/Raidark/shared/api/driver/modules"
	domprovider "github.com/r0x16/Raidark/shared/providers/domain"
	driverprovider "github.com/r0x16/Raidark/shared/providers/driver"
)

func main() {

	raidark := raidark.New(getProviders())
	modules := getModules(raidark)
	raidark.Run(modules)
}

func getModules(raidark *raidark.Raidark) []apidomain.ApiModule {

	authRoot := raidark.RootModule("/auth")
	apiv1Root := raidark.AuthenticatedRootModule("/api/v1")

	return []apidomain.ApiModule{
		&moduleapi.EchoAuthModule{EchoModule: authRoot},
		&moduleapi.EchoApiMainModule{EchoModule: apiv1Root},
	}
}

func getProviders() []domprovider.ProviderFactory {
	return []domprovider.ProviderFactory{
		&driverprovider.DatastoreProviderFactory{},
		&driverprovider.AuthProviderFactory{},
		&driverprovider.ApiProviderFactory{},
	}
}

package controller

import (
	driverapi "github.com/r0x16/Raidark/shared/api/driver"
	driverdatastore "github.com/r0x16/Raidark/shared/datastore/driver"
)

type DbMigrationController struct {
	*driverapi.ApplicationBundle
}

func (d *DbMigrationController) MigrateAction() error {
	modelsError := d.migrateModels()
	if modelsError != nil {
		return modelsError
	}

	return nil
}

func (d *DbMigrationController) migrateModels() error {
	db := d.Database.(*driverdatastore.GormPostgresDatabaseProvider)
	return db.GetDataStore().Exec.AutoMigrate()
}

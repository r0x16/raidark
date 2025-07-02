package controller

import (
	"github.com/r0x16/Raidark/api/drivers"
	"github.com/r0x16/Raidark/shared/driver/db"
)

type DbMigrationController struct {
	*drivers.ApplicationBundle
}

func (d *DbMigrationController) MigrateAction() error {
	modelsError := d.migrateModels()
	if modelsError != nil {
		return modelsError
	}

	return nil
}

func (d *DbMigrationController) migrateModels() error {
	db := d.Database.(*db.GormPostgresDatabaseProvider)
	return db.Connection.AutoMigrate()
}

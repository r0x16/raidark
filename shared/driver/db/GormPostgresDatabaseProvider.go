package db

import (
	"os"

	"github.com/r0x16/Raidark/shared/domain"
	"github.com/r0x16/Raidark/shared/domain/model/db"
	"github.com/r0x16/Raidark/shared/driver/db/connection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

/*
 * Represents a postgres database provider connector using gorm
 */
type GormPostgresDatabaseProvider struct {
	Datastore *db.DataStore
}

var _ domain.DatabaseProvider = &GormPostgresDatabaseProvider{}

/*
 * Creates a new dsn string for the postgres driver
 * using the connection struct and the environment variables
 */
func (g *GormPostgresDatabaseProvider) Connect() error {
	dsn := connection.GormPostgresConnection{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
	}

	var err error
	connection, err := gorm.Open(postgres.Open(dsn.GetDsn()), &gorm.Config{})
	if err != nil {
		return err
	}

	g.Datastore = db.NewDataStore(connection)
	return nil
}

/*
 * Close the database connection
 * This method ensures that the underlying SQL database connection is properly closed.
 */
func (g *GormPostgresDatabaseProvider) Close() error {
	if g.Datastore != nil {
		sqlDB, err := g.Datastore.Exec.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func (g *GormPostgresDatabaseProvider) GetDataStore() *db.DataStore {
	return g.Datastore
}

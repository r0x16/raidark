package driver

import (
	"github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/r0x16/Raidark/shared/datastore/driver/connection"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Represents a postgres database provider connector using gorm
type GormPostgresDatabaseProvider struct {
	db          *gorm.DB
	envProvider domenv.EnvProvider

	// Deprecated: Use GetTransaction() instead
	Datastore *domain.DataStore
}

var _ domain.DatabaseProvider = &GormPostgresDatabaseProvider{}

// NewGormPostgresDatabaseProvider creates a new postgres database provider with EnvProvider
func NewGormPostgresDatabaseProvider(envProvider domenv.EnvProvider) *GormPostgresDatabaseProvider {
	return &GormPostgresDatabaseProvider{
		envProvider: envProvider,
	}
}

// Creates a new dsn string for the postgres driver
// using the connection struct and the environment variables with defaults
func (g *GormPostgresDatabaseProvider) Connect() error {
	dsn := connection.GormPostgresConnection{
		Host:     g.envProvider.GetString("DB_HOST", "localhost"),
		Port:     g.envProvider.GetString("DB_PORT", "5432"),
		Username: g.envProvider.GetString("DB_USER", "raidark"),
		Password: g.envProvider.GetString("DB_PASSWORD", ""),
		Database: g.envProvider.GetString("DB_DATABASE", "raidark"),
	}

	var err error
	connection, err := gorm.Open(postgres.Open(dsn.GetDsn()), &gorm.Config{})
	if err != nil {
		return err
	}

	g.db = connection
	g.Datastore = domain.NewDataStore(connection)
	return nil
}

// Close the database connection
// This method ensures that the underlying SQL database connection is properly closed.
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

// Deprecated: Use GetTransaction() instead
func (g *GormPostgresDatabaseProvider) GetDataStore() *domain.DataStore {
	return g.Datastore
}

func (g *GormPostgresDatabaseProvider) GetTransaction() domain.Transaction {
	return NewGormTransaction(g.db)
}

package driver

import (
	"github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/r0x16/Raidark/shared/datastore/driver/connection"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Represents a mysql database provider connector using gorm
type GormMysqlDatabaseProvider struct {
	db          *gorm.DB
	envProvider domenv.EnvProvider

	// Deprecated: Use GetTransaction() instead
	Datastore *domain.DataStore
}

var _ domain.DatabaseProvider = &GormMysqlDatabaseProvider{}

// NewGormMysqlDatabaseProvider creates a new mysql database provider with EnvProvider
func NewGormMysqlDatabaseProvider(envProvider domenv.EnvProvider) *GormMysqlDatabaseProvider {
	return &GormMysqlDatabaseProvider{
		envProvider: envProvider,
	}
}

// Creates a new dsn string for the mysql driver
// using the connection struct and the environment variables with defaults
func (g *GormMysqlDatabaseProvider) Connect() error {
	dsn := connection.GormMysqlConnection{
		Host:     g.envProvider.GetString("DB_HOST", "localhost"),
		Port:     g.envProvider.GetString("DB_PORT", "3306"),
		Username: g.envProvider.GetString("DB_USER", "raidark"),
		Password: g.envProvider.GetString("DB_PASSWORD", ""),
		Database: g.envProvider.GetString("DB_DATABASE", "raidark"),
	}

	var err error
	connection, err := gorm.Open(mysql.Open(dsn.GetDsn()), &gorm.Config{})
	if err != nil {
		return err
	}

	g.db = connection
	g.Datastore = domain.NewDataStore(connection)
	return err
}

// Close the database connection
// This method ensures that the underlying SQL database connection is properly closed.
func (g *GormMysqlDatabaseProvider) Close() error {
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
func (g *GormMysqlDatabaseProvider) GetDataStore() *domain.DataStore {
	return g.Datastore
}

func (g *GormMysqlDatabaseProvider) GetTransaction() domain.Transaction {
	return NewGormTransaction(g.db)
}

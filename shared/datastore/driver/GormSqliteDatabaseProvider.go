package driver

import (
	"github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/r0x16/Raidark/shared/datastore/driver/connection"
	domenv "github.com/r0x16/Raidark/shared/env/domain"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// Represents a sqlite database provider connector using gorm
type GormSqliteDatabaseProvider struct {
	db          *gorm.DB
	envProvider domenv.EnvProvider

	// Deprecated: Use GetTransaction() instead
	Datastore *domain.DataStore
}

var _ domain.DatabaseProvider = &GormSqliteDatabaseProvider{}

// NewGormSqliteDatabaseProvider creates a new sqlite database provider with EnvProvider
func NewGormSqliteDatabaseProvider(envProvider domenv.EnvProvider) *GormSqliteDatabaseProvider {
	return &GormSqliteDatabaseProvider{
		envProvider: envProvider,
	}
}

// Creates a new dsn string for the sqlite driver
// using the connection struct and the environment variables with defaults
func (g *GormSqliteDatabaseProvider) Connect() error {
	dsn := connection.GormSqliteConnection{
		DatabasePath: g.envProvider.GetString("DB_DATABASE", "raidark.db"),
	}

	var err error
	connection, err := gorm.Open(sqlite.Open(dsn.GetDsn()), &gorm.Config{})
	if err != nil {
		return err
	}

	g.db = connection
	g.Datastore = domain.NewDataStore(connection)
	return nil
}

// Close the database connection
// This method ensures that the underlying SQL database connection is properly closed.
func (g *GormSqliteDatabaseProvider) Close() error {
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
func (g *GormSqliteDatabaseProvider) GetDataStore() *domain.DataStore {
	return g.Datastore
}

func (g *GormSqliteDatabaseProvider) GetTransaction() domain.Transaction {
	return NewGormTransaction(g.db)
}

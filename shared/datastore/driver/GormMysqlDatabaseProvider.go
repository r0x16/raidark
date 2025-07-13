package driver

import (
	"os"

	"github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/r0x16/Raidark/shared/datastore/driver/connection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
 * Represents a mysql database provider connector using gorm
 */
type GormMysqlDatabaseProvider struct {
	db *gorm.DB

	// Deprecated: Use GetTransaction() instead
	Datastore *domain.DataStore
}

var _ domain.DatabaseProvider = &GormMysqlDatabaseProvider{}

/*
 * Creates a new dsn string for the mysql driver
 * using the connection struct and the environment variables
 */
func (g *GormMysqlDatabaseProvider) Connect() error {
	dsn := connection.GormMysqlConnection{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Username: os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Database: os.Getenv("DB_DATABASE"),
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

/*
 * Close the database connection
 * This method ensures that the underlying SQL database connection is properly closed.
 */
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

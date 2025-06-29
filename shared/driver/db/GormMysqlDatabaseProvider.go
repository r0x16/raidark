package db

import (
	"os"

	"github.com/r0x16/Raidark/shared/domain"
	"github.com/r0x16/Raidark/shared/driver/db/connection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

/*
 * Represents a mysql database provider connector using gorm
 */
type GormMysqlDatabaseProvider struct {
	Connection *gorm.DB
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
	g.Connection, err = gorm.Open(mysql.Open(dsn.GetDsn()), &gorm.Config{})
	return err
}

/*
 * Close the database connection
 * This method ensures that the underlying SQL database connection is properly closed.
 */
func (g *GormMysqlDatabaseProvider) Close() error {
	if g.Connection != nil {
		sqlDB, err := g.Connection.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

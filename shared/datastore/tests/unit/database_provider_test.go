package unit

import (
	"os"
	"testing"

	"github.com/r0x16/Raidark/shared/datastore/driver"
	"github.com/r0x16/Raidark/shared/datastore/driver/connection"
	"github.com/stretchr/testify/assert"
)

// TestGormMysqlDatabaseProviderInstantiation tests MySQL provider creation
func TestGormMysqlDatabaseProviderInstantiation(t *testing.T) {
	provider := &driver.GormMysqlDatabaseProvider{}
	assert.NotNil(t, provider, "MySQL provider should be instantiable")
	assert.Nil(t, provider.Datastore, "Datastore should be nil before connection")
}

// TestGormPostgresDatabaseProviderInstantiation tests PostgreSQL provider creation
func TestGormPostgresDatabaseProviderInstantiation(t *testing.T) {
	provider := &driver.GormPostgresDatabaseProvider{}
	assert.NotNil(t, provider, "PostgreSQL provider should be instantiable")
	assert.Nil(t, provider.Datastore, "Datastore should be nil before connection")
}

// TestMysqlConnectionDsnGeneration tests MySQL DSN string generation
func TestMysqlConnectionDsnGeneration(t *testing.T) {
	conn := connection.GormMysqlConnection{
		Host:     "localhost",
		Port:     "3306",
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
	}

	expectedDsn := "testuser:testpass@tcp(localhost:3306)/testdb?charset=utf8mb4&parseTime=True&loc=Local"
	actualDsn := conn.GetDsn()
	
	assert.Equal(t, expectedDsn, actualDsn, "MySQL DSN should match expected format")
}

// TestPostgresConnectionDsnGeneration tests PostgreSQL DSN string generation
func TestPostgresConnectionDsnGeneration(t *testing.T) {
	conn := connection.GormPostgresConnection{
		Host:     "localhost",
		Port:     "5432",
		Username: "testuser",
		Password: "testpass",
		Database: "testdb",
	}

	expectedDsn := "host='localhost' user='testuser' password='testpass' dbname='testdb' port='5432' sslmode='disable'"
	actualDsn := conn.GetDsn()
	
	assert.Equal(t, expectedDsn, actualDsn, "PostgreSQL DSN should match expected format")
}

// TestMysqlConnectionWithEmptyValues tests MySQL connection with empty values
func TestMysqlConnectionWithEmptyValues(t *testing.T) {
	conn := connection.GormMysqlConnection{}
	dsn := conn.GetDsn()
	
	// Should generate DSN even with empty values
	expectedDsn := ":@tcp(:)/?charset=utf8mb4&parseTime=True&loc=Local"
	assert.Equal(t, expectedDsn, dsn, "Should handle empty values gracefully")
}

// TestPostgresConnectionWithEmptyValues tests PostgreSQL connection with empty values
func TestPostgresConnectionWithEmptyValues(t *testing.T) {
	conn := connection.GormPostgresConnection{}
	dsn := conn.GetDsn()
	
	// Should generate DSN even with empty values
	expectedDsn := "host='' user='' password='' dbname='' port='' sslmode='disable'"
	assert.Equal(t, expectedDsn, dsn, "Should handle empty values gracefully")
}

// TestMysqlProviderWithoutConnection tests MySQL provider behavior before connection
func TestMysqlProviderWithoutConnection(t *testing.T) {
	provider := &driver.GormMysqlDatabaseProvider{}
	
	// Test GetDataStore before connection
	datastore := provider.GetDataStore()
	assert.Nil(t, datastore, "Should return nil datastore before connection")
	
	// Test Close before connection
	err := provider.Close()
	assert.NoError(t, err, "Should not error when closing unconnected provider")
}

// TestPostgresProviderWithoutConnection tests PostgreSQL provider behavior before connection
func TestPostgresProviderWithoutConnection(t *testing.T) {
	provider := &driver.GormPostgresDatabaseProvider{}
	
	// Test GetDataStore before connection
	datastore := provider.GetDataStore()
	assert.Nil(t, datastore, "Should return nil datastore before connection")
	
	// Test Close before connection
	err := provider.Close()
	assert.NoError(t, err, "Should not error when closing unconnected provider")
}

// TestMysqlProviderConnectionWithInvalidCredentials tests MySQL provider with invalid credentials
func TestMysqlProviderConnectionWithInvalidCredentials(t *testing.T) {
	// Set invalid environment variables
	originalValues := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_DATABASE": os.Getenv("DB_DATABASE"),
	}
	
	// Set test environment variables
	os.Setenv("DB_HOST", "invalid_host")
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_USER", "invalid_user")
	os.Setenv("DB_PASSWORD", "invalid_pass")
	os.Setenv("DB_DATABASE", "invalid_db")
	
	defer func() {
		// Restore original environment variables
		for key, value := range originalValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	provider := &driver.GormMysqlDatabaseProvider{}
	err := provider.Connect()
	
	// Should return error with invalid credentials
	assert.Error(t, err, "Should return error with invalid database credentials")
	assert.Nil(t, provider.Datastore, "Datastore should remain nil on connection failure")
}

// TestPostgresProviderConnectionWithInvalidCredentials tests PostgreSQL provider with invalid credentials
func TestPostgresProviderConnectionWithInvalidCredentials(t *testing.T) {
	// Set invalid environment variables
	originalValues := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_DATABASE": os.Getenv("DB_DATABASE"),
	}
	
	// Set test environment variables
	os.Setenv("DB_HOST", "invalid_host")
	os.Setenv("DB_PORT", "1234")
	os.Setenv("DB_USER", "invalid_user")
	os.Setenv("DB_PASSWORD", "invalid_pass")
	os.Setenv("DB_DATABASE", "invalid_db")
	
	defer func() {
		// Restore original environment variables
		for key, value := range originalValues {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	provider := &driver.GormPostgresDatabaseProvider{}
	err := provider.Connect()
	
	// Should return error with invalid credentials
	assert.Error(t, err, "Should return error with invalid database credentials")
	assert.Nil(t, provider.Datastore, "Datastore should remain nil on connection failure")
}
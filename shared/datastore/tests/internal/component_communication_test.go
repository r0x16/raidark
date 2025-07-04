package internal

import (
	"os"
	"testing"

	"github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/r0x16/Raidark/shared/datastore/driver"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// ComponentCommunicationTestSuite tests how components communicate with each other
type ComponentCommunicationTestSuite struct {
	suite.Suite
	mockDB       *gorm.DB
	datastore    *domain.DataStore
	mysqlProvider *driver.GormMysqlDatabaseProvider
	pgProvider    *driver.GormPostgresDatabaseProvider
}

// SetupSuite sets up the test suite
func (suite *ComponentCommunicationTestSuite) SetupSuite() {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)
	
	suite.mockDB = db
	suite.datastore = domain.NewDataStore(db)
	suite.mysqlProvider = &driver.GormMysqlDatabaseProvider{}
	suite.pgProvider = &driver.GormPostgresDatabaseProvider{}
}

// TestDataStoreProviderIntegration tests integration between providers and datastore
func (suite *ComponentCommunicationTestSuite) TestDataStoreProviderIntegration() {
	// Test MySQL provider datastore assignment
	suite.mysqlProvider.Datastore = suite.datastore
	retrievedDatastore := suite.mysqlProvider.GetDataStore()
	
	suite.Assert().NotNil(retrievedDatastore)
	suite.Assert().Equal(suite.datastore, retrievedDatastore)
	suite.Assert().Equal(suite.mockDB, retrievedDatastore.Exec)
	
	// Test PostgreSQL provider datastore assignment
	suite.pgProvider.Datastore = suite.datastore
	retrievedDatastorePg := suite.pgProvider.GetDataStore()
	
	suite.Assert().NotNil(retrievedDatastorePg)
	suite.Assert().Equal(suite.datastore, retrievedDatastorePg)
	suite.Assert().Equal(suite.mockDB, retrievedDatastorePg.Exec)
}

// TestProviderLifecycleManagement tests the lifecycle of database providers
func (suite *ComponentCommunicationTestSuite) TestProviderLifecycleManagement() {
	// Create fresh providers for lifecycle testing
	freshMysqlProvider := &driver.GormMysqlDatabaseProvider{}
	freshPgProvider := &driver.GormPostgresDatabaseProvider{}
	
	// Test provider initialization state
	suite.Assert().Nil(freshMysqlProvider.Datastore)
	suite.Assert().Nil(freshPgProvider.Datastore)
	
	// Simulate connection by setting datastore
	freshMysqlProvider.Datastore = suite.datastore
	freshPgProvider.Datastore = suite.datastore
	
	// Test provider state after connection
	suite.Assert().NotNil(freshMysqlProvider.Datastore)
	suite.Assert().NotNil(freshPgProvider.Datastore)
	
	// Test close operations
	err := freshMysqlProvider.Close()
	suite.Assert().NoError(err)
	
	err = freshPgProvider.Close()
	suite.Assert().NoError(err)
}

// TestMultipleProvidersSameDatastore tests multiple providers sharing the same datastore
func (suite *ComponentCommunicationTestSuite) TestMultipleProvidersSameDatastore() {
	sharedDatastore := domain.NewDataStore(suite.mockDB)
	
	// Assign same datastore to both providers
	suite.mysqlProvider.Datastore = sharedDatastore
	suite.pgProvider.Datastore = sharedDatastore
	
	// Verify both providers reference the same datastore
	mysqlDS := suite.mysqlProvider.GetDataStore()
	pgDS := suite.pgProvider.GetDataStore()
	
	suite.Assert().Equal(mysqlDS, pgDS)
	suite.Assert().Equal(sharedDatastore, mysqlDS)
	suite.Assert().Equal(sharedDatastore, pgDS)
	
	// Verify they share the same underlying database connection
	suite.Assert().Equal(mysqlDS.Exec, pgDS.Exec)
}

// TestDataStoreOperationThroughProvider tests database operations through provider
func (suite *ComponentCommunicationTestSuite) TestDataStoreOperationThroughProvider() {
	// Create test model
	type TestModel struct {
		domain.BaseModel
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	
	// Migrate the test model
	err := suite.mockDB.AutoMigrate(&TestModel{})
	suite.Require().NoError(err)
	
	// Assign datastore to provider
	suite.mysqlProvider.Datastore = suite.datastore
	
	// Get datastore through provider
	ds := suite.mysqlProvider.GetDataStore()
	suite.Require().NotNil(ds)
	
	// Perform CRUD operations through provider's datastore
	testRecord := TestModel{
		Name:        "Test Record",
		Description: "Test Description",
	}
	
	// Create
	result := ds.Exec.Create(&testRecord)
	suite.Assert().NoError(result.Error)
	suite.Assert().Equal(int64(1), result.RowsAffected)
	suite.Assert().NotEqual(uint(0), testRecord.ID)
	
	// Read
	var retrievedRecord TestModel
	result = ds.Exec.First(&retrievedRecord, testRecord.ID)
	suite.Assert().NoError(result.Error)
	suite.Assert().Equal(testRecord.Name, retrievedRecord.Name)
	suite.Assert().Equal(testRecord.Description, retrievedRecord.Description)
	
	// Update
	retrievedRecord.Description = "Updated Description"
	result = ds.Exec.Save(&retrievedRecord)
	suite.Assert().NoError(result.Error)
	
	// Verify update
	var updatedRecord TestModel
	result = ds.Exec.First(&updatedRecord, testRecord.ID)
	suite.Assert().NoError(result.Error)
	suite.Assert().Equal("Updated Description", updatedRecord.Description)
	
	// Delete
	result = ds.Exec.Delete(&updatedRecord)
	suite.Assert().NoError(result.Error)
	
	// Verify deletion
	var deletedRecord TestModel
	result = ds.Exec.First(&deletedRecord, testRecord.ID)
	suite.Assert().Error(result.Error)
}

// TestProviderInterfaceCompliance tests that providers implement the interface correctly
func (suite *ComponentCommunicationTestSuite) TestProviderInterfaceCompliance() {
	// Test that providers implement the DatabaseProvider interface
	var mysqlProvider domain.DatabaseProvider = suite.mysqlProvider
	var pgProvider domain.DatabaseProvider = suite.pgProvider
	
	suite.Assert().NotNil(mysqlProvider)
	suite.Assert().NotNil(pgProvider)
	
	// Test interface methods exist and can be called
	suite.Assert().NotPanics(func() {
		mysqlProvider.GetDataStore()
		mysqlProvider.Close()
	})
	
	suite.Assert().NotPanics(func() {
		pgProvider.GetDataStore()
		pgProvider.Close()
	})
}

// TestEnvironmentVariableInteraction tests how providers interact with environment variables
func (suite *ComponentCommunicationTestSuite) TestEnvironmentVariableInteraction() {
	// Store original environment variables
	originalVars := map[string]string{
		"DB_HOST":     os.Getenv("DB_HOST"),
		"DB_PORT":     os.Getenv("DB_PORT"),
		"DB_USER":     os.Getenv("DB_USER"),
		"DB_PASSWORD": os.Getenv("DB_PASSWORD"),
		"DB_DATABASE": os.Getenv("DB_DATABASE"),
	}
	
	// Set test environment variables
	testVars := map[string]string{
		"DB_HOST":     "test_host",
		"DB_PORT":     "3306",
		"DB_USER":     "test_user",
		"DB_PASSWORD": "test_pass",
		"DB_DATABASE": "test_db",
	}
	
	for key, value := range testVars {
		os.Setenv(key, value)
	}
	
	defer func() {
		// Restore original environment variables
		for key, value := range originalVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()
	
	// Test that providers attempt to use environment variables during connection
	// Note: These will fail to connect but should read the environment variables
	mysqlProvider := &driver.GormMysqlDatabaseProvider{}
	pgProvider := &driver.GormPostgresDatabaseProvider{}
	
	// Both should attempt to connect and fail (but not panic)
	err := mysqlProvider.Connect()
	suite.Assert().Error(err) // Should fail with invalid credentials but not panic
	
	err = pgProvider.Connect()
	suite.Assert().Error(err) // Should fail with invalid credentials but not panic
}

// TestConcurrentProviderAccess tests concurrent access to providers
func (suite *ComponentCommunicationTestSuite) TestConcurrentProviderAccess() {
	sharedDatastore := domain.NewDataStore(suite.mockDB)
	suite.mysqlProvider.Datastore = sharedDatastore
	
	// Channel to collect results from goroutines
	results := make(chan *domain.DataStore, 10)
	
	// Start multiple goroutines accessing the same provider
	for i := 0; i < 10; i++ {
		go func() {
			ds := suite.mysqlProvider.GetDataStore()
			results <- ds
		}()
	}
	
	// Collect results
	for i := 0; i < 10; i++ {
		ds := <-results
		suite.Assert().Equal(sharedDatastore, ds)
	}
}

// Run the test suite
func TestComponentCommunicationTestSuite(t *testing.T) {
	suite.Run(t, new(ComponentCommunicationTestSuite))
}
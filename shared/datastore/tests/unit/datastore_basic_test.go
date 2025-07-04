package unit

import (
	"testing"

	"github.com/r0x16/Raidark/shared/datastore/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TestDataStoreInstantiation tests basic DataStore creation
func TestDataStoreInstantiation(t *testing.T) {
	// Create in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err, "Should create test database without error")

	// Test DataStore creation
	datastore := domain.NewDataStore(db)
	assert.NotNil(t, datastore, "DataStore should not be nil")
	assert.NotNil(t, datastore.Exec, "DataStore.Exec should not be nil")
	assert.Equal(t, db, datastore.Exec, "DataStore.Exec should match provided DB")
}

// TestDataStoreWithNilDB tests DataStore behavior with nil database
func TestDataStoreWithNilDB(t *testing.T) {
	datastore := domain.NewDataStore(nil)
	assert.NotNil(t, datastore, "DataStore should not be nil even with nil DB")
	assert.Nil(t, datastore.Exec, "DataStore.Exec should be nil when created with nil DB")
}

// TestBaseModelStructure tests BaseModel basic structure
func TestBaseModelStructure(t *testing.T) {
	model := domain.BaseModel{}
	
	// Test that BaseModel can be instantiated
	assert.NotNil(t, &model, "BaseModel should be instantiable")
	
	// Test zero values
	assert.Equal(t, uint(0), model.ID, "ID should have zero value")
	assert.True(t, model.CreatedAt.IsZero(), "CreatedAt should have zero value")
	assert.True(t, model.UpdatedAt.IsZero(), "UpdatedAt should have zero value")
}

// TestDataStoreOperationWithMockDB tests basic operations with mock database
func TestDataStoreOperationWithMockDB(t *testing.T) {
	// Create in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err, "Should create test database without error")

	// Create sample model for testing
	type TestModel struct {
		domain.BaseModel
		Name string `json:"name"`
	}

	// Auto-migrate the test model
	err = db.AutoMigrate(&TestModel{})
	assert.NoError(t, err, "Should auto-migrate test model without error")

	// Create DataStore
	datastore := domain.NewDataStore(db)

	// Test basic database operation
	testRecord := TestModel{Name: "Test Record"}
	result := datastore.Exec.Create(&testRecord)
	assert.NoError(t, result.Error, "Should create record without error")
	assert.Equal(t, int64(1), result.RowsAffected, "Should affect one row")
	assert.NotEqual(t, uint(0), testRecord.ID, "Record should have assigned ID")
}

// TestDataStoreConnectionStatus tests if DataStore can check connection status
func TestDataStoreConnectionStatus(t *testing.T) {
	// Create in-memory database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err, "Should create test database without error")

	datastore := domain.NewDataStore(db)

	// Test connection status through raw SQL
	var result int
	err = datastore.Exec.Raw("SELECT 1").Scan(&result).Error
	assert.NoError(t, err, "Should execute raw SQL without error")
	assert.Equal(t, 1, result, "Should return expected result")
}
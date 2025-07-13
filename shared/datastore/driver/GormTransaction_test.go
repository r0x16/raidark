package driver

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

func newDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	return db
}

func TestTransactionLifecycle(t *testing.T) {
	tx := NewGormTransaction(newDB(t))
	tx.Begin()
	if !tx.IsInTransaction {
		t.Fatal("expected begin")
	}
	tx.Commit()
	if tx.IsInTransaction {
		t.Fatal("expected committed")
	}
	tx.Begin()
	tx.Rollback()
	if tx.IsInTransaction {
		t.Fatal("expected rolled back")
	}
}

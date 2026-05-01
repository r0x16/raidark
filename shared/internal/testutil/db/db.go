// Package db contiene helpers de bases de datos efímeras para tests.
package db

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// NewSQLite abre una base SQLite en memoria para tests, registra el cleanup de
// la conexión y aplica AutoMigrate cuando se entregan modelos.
func NewSQLite(t testing.TB, models ...any) *gorm.DB {
	t.Helper()

	database, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite test database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("unwrap sqlite test database: %v", err)
	}
	t.Cleanup(func() {
		if err := sqlDB.Close(); err != nil {
			t.Errorf("close sqlite test database: %v", err)
		}
	})

	if len(models) > 0 {
		if err := database.AutoMigrate(models...); err != nil {
			t.Fatalf("auto migrate sqlite test database: %v", err)
		}
	}

	return database
}

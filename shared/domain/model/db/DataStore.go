package db

import "gorm.io/gorm"

type DataStore struct {
	Exec *gorm.DB
}

func NewDataStore(db *gorm.DB) *DataStore {
	return &DataStore{Exec: db}
}

package db

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type smokeModel struct {
	ID   uint `gorm:"primarykey"`
	Name string
}

func TestNewSQLite_smoke(t *testing.T) {
	database := NewSQLite(t, &smokeModel{})

	require.NoError(t, database.Create(&smokeModel{Name: "raidark"}).Error)

	var count int64
	require.NoError(t, database.Model(&smokeModel{}).Where("name = ?", "raidark").Count(&count).Error)
	assert.Equal(t, int64(1), count)
}

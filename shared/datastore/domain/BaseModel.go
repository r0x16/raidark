package domain

import (
	"time"

	"github.com/r0x16/Raidark/shared/ids"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        ids.UUIDv7     `gorm:"primarykey;type:varchar(36)"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		id, err := ids.NewV7()
		if err != nil {
			return err
		}
		m.ID = ids.UUIDv7(id)
	}
	return nil
}

package model

import (
	"time"

	"github.com/linkyfish/kxl_backend_go/internal/util"
	"gorm.io/gorm"
)

// UUIDModel matches tables that use UUID as primary key and have created_at/updated_at.
type UUIDModel struct {
	ID        string    `gorm:"type:uuid;primaryKey;column:id" json:"id"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (m *UUIDModel) BeforeCreate(tx *gorm.DB) error {
	if m.ID == "" {
		m.ID = util.NewUUID()
	}
	return nil
}

type CreatedAtOnly struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

type Timestamps struct {
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}


package model

import "time"

type ProjectFeature struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ProjectID   string    `gorm:"type:uuid;column:project_id" json:"project_id"`
	Name        string    `gorm:"column:name" json:"name"`
	Description string    `gorm:"column:description" json:"description"`
	Icon        *string   `gorm:"column:icon" json:"icon"`
	SortOrder   int       `gorm:"column:sort_order" json:"sort_order"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProjectFeature) TableName() string { return "project_features" }


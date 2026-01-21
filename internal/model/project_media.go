package model

import "time"

type ProjectMedia struct {
	ID        int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ProjectID string    `gorm:"type:uuid;column:project_id" json:"project_id"`
	Type      string    `gorm:"column:type" json:"type"`
	URL       string    `gorm:"column:url" json:"url"`
	Title     *string   `gorm:"column:title" json:"title"`
	SortOrder int       `gorm:"column:sort_order" json:"sort_order"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProjectMedia) TableName() string { return "project_media" }


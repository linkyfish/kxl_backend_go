package model

import "time"

type ProjectVersion struct {
	ID          int       `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	ProjectID   string    `gorm:"type:uuid;column:project_id" json:"project_id"`
	Version     string    `gorm:"column:version" json:"version"`
	ReleaseDate time.Time `gorm:"column:release_date" json:"release_date"`
	Changelog   string    `gorm:"column:changelog" json:"changelog"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ProjectVersion) TableName() string { return "project_versions" }


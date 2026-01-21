package model

type ProjectTag struct {
	ProjectID string `gorm:"type:uuid;primaryKey;column:project_id" json:"project_id"`
	TagID     int    `gorm:"primaryKey;column:tag_id" json:"tag_id"`
}

func (ProjectTag) TableName() string { return "project_tags" }


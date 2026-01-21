package model

type CaseProject struct {
	CaseID    string `gorm:"type:uuid;primaryKey;column:case_id" json:"case_id"`
	ProjectID string `gorm:"type:uuid;primaryKey;column:project_id" json:"project_id"`
}

func (CaseProject) TableName() string { return "case_projects" }


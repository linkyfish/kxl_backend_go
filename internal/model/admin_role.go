package model

type AdminRole struct {
	Code        string `gorm:"primaryKey;column:code" json:"code"`
	Name        string `gorm:"column:name" json:"name"`
	Description string `gorm:"column:description" json:"description"`
	IsSystem    bool   `gorm:"column:is_system" json:"is_system"`
	Timestamps
}

func (AdminRole) TableName() string { return "admin_roles" }


package model

type AdminPermission struct {
	Code        string `gorm:"primaryKey;column:code" json:"code"`
	Name        string `gorm:"column:name" json:"name"`
	GroupName   string `gorm:"column:group_name" json:"group_name"`
	Description string `gorm:"column:description" json:"description"`
	IsSystem    bool   `gorm:"column:is_system" json:"is_system"`
	Timestamps
}

func (AdminPermission) TableName() string { return "admin_permissions" }


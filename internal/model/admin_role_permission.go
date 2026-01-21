package model

import "time"

type AdminRolePermission struct {
	RoleCode       string    `gorm:"primaryKey;column:role_code" json:"role_code"`
	PermissionCode string    `gorm:"primaryKey;column:permission_code" json:"permission_code"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (AdminRolePermission) TableName() string { return "admin_role_permissions" }


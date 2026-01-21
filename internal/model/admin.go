package model

type Admin struct {
	UUIDModel

	Username     string `gorm:"column:username" json:"username"`
	PasswordHash string `gorm:"column:password_hash" json:"-"`
	Role         string `gorm:"column:role" json:"role"`
	Status       int16  `gorm:"column:status" json:"status"`
}

func (Admin) TableName() string { return "admins" }


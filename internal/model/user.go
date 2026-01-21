package model

type User struct {
	UUIDModel

	Username       string `gorm:"column:username" json:"username"`
	Email          string `gorm:"column:email" json:"email"`
	PasswordHash   string `gorm:"column:password_hash" json:"-"`
	Status         int16  `gorm:"column:status" json:"status"`
	SessionVersion int    `gorm:"column:session_version" json:"session_version"`
}

func (User) TableName() string { return "users" }


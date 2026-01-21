package model

type Message struct {
	ID      int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name    string `gorm:"column:name" json:"name"`
	Company *string `gorm:"column:company" json:"company"`
	Phone   string `gorm:"column:phone" json:"phone"`
	Email   string `gorm:"column:email" json:"email"`
	Content string `gorm:"column:content" json:"content"`
	Status  int16  `gorm:"column:status" json:"status"`
	Note    *string `gorm:"column:note" json:"note"`
	Timestamps
}

func (Message) TableName() string { return "messages" }


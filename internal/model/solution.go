package model

type Solution struct {
	ID          int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string  `gorm:"column:name" json:"name"`
	Description string  `gorm:"column:description" json:"description"`
	Icon        *string `gorm:"column:icon" json:"icon"`
	BgClass     string  `gorm:"column:bg_class" json:"bg_class"`
	Link        string  `gorm:"column:link" json:"link"`
	SortOrder   int     `gorm:"column:sort_order" json:"sort_order"`
	IsVisible   bool    `gorm:"column:is_visible" json:"is_visible"`
	Timestamps
}

func (Solution) TableName() string { return "solutions" }


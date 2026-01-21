package model

type Partner struct {
	ID        int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string  `gorm:"column:name" json:"name"`
	Logo      *string `gorm:"column:logo" json:"logo"`
	Website   *string `gorm:"column:website" json:"website"`
	SortOrder int     `gorm:"column:sort_order" json:"sort_order"`
	IsVisible bool    `gorm:"column:is_visible" json:"is_visible"`
	Timestamps
}

func (Partner) TableName() string { return "partners" }


package model

type FriendlyLink struct {
	ID          int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name        string  `gorm:"column:name" json:"name"`
	URL         string  `gorm:"column:url" json:"url"`
	Logo        *string `gorm:"column:logo" json:"logo"`
	Description *string `gorm:"column:description" json:"description"`
	SortOrder   int     `gorm:"column:sort_order" json:"sort_order"`
	IsVisible   bool    `gorm:"column:is_visible" json:"is_visible"`
	Timestamps
}

func (FriendlyLink) TableName() string { return "friendly_links" }


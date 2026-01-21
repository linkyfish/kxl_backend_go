package model

type Banner struct {
	ID        int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Title     string  `gorm:"column:title" json:"title"`
	Subtitle  *string `gorm:"column:subtitle" json:"subtitle"`
	Highlight *string `gorm:"column:highlight" json:"highlight"`
	Tag       *string `gorm:"column:tag" json:"tag"`
	Image     *string `gorm:"column:image" json:"image"`
	Link      *string `gorm:"column:link" json:"link"`
	LinkText  string  `gorm:"column:link_text" json:"link_text"`
	BgClass   string  `gorm:"column:bg_class" json:"bg_class"`
	SortOrder int     `gorm:"column:sort_order" json:"sort_order"`
	IsVisible bool    `gorm:"column:is_visible" json:"is_visible"`
	Timestamps
}

func (Banner) TableName() string { return "banners" }


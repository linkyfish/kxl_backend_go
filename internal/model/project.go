package model

type Project struct {
	UUIDModel

	Name        string  `gorm:"column:name" json:"name"`
	Description string  `gorm:"column:description" json:"description"`
	CoverImage  *string `gorm:"column:cover_image" json:"cover_image"`
	CategoryID  *int    `gorm:"column:category_id" json:"category_id"`
	Status      int16   `gorm:"column:status" json:"status"`
	SortOrder   int     `gorm:"column:sort_order" json:"sort_order"`
}

func (Project) TableName() string { return "projects" }


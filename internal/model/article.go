package model

import "time"

type Article struct {
	UUIDModel

	Title       string     `gorm:"column:title" json:"title"`
	Summary     string     `gorm:"column:summary" json:"summary"`
	Content     string     `gorm:"column:content" json:"content"`
	CoverImage  *string    `gorm:"column:cover_image" json:"cover_image"`
	CategoryID  *int       `gorm:"column:category_id" json:"category_id"`
	ViewCount   int        `gorm:"column:view_count" json:"view_count"`
	Status      int16      `gorm:"column:status" json:"status"`
	PublishedAt *time.Time `gorm:"column:published_at" json:"published_at"`
}

func (Article) TableName() string { return "articles" }


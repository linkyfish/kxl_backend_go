package model

import "gorm.io/datatypes"

// CaseStudy maps to table `cases` (Go reserved word: case).
type CaseStudy struct {
	UUIDModel

	ClientName        string         `gorm:"column:client_name" json:"client_name"`
	CoverImage        *string        `gorm:"column:cover_image" json:"cover_image"`
	Summary           string         `gorm:"column:summary" json:"summary"`
	Background        string         `gorm:"column:background" json:"background"`
	Solution          string         `gorm:"column:solution" json:"solution"`
	Results           datatypes.JSON `gorm:"type:jsonb;column:results" json:"results"`
	Testimonial       *string        `gorm:"column:testimonial" json:"testimonial"`
	TestimonialAuthor *string        `gorm:"column:testimonial_author" json:"testimonial_author"`
	TestimonialTitle  *string        `gorm:"column:testimonial_title" json:"testimonial_title"`
	CategoryID        *int           `gorm:"column:category_id" json:"category_id"`
	Status            int16          `gorm:"column:status" json:"status"`
}

func (CaseStudy) TableName() string { return "cases" }

package model

type Testimonial struct {
	ID        int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string  `gorm:"column:name" json:"name"`
	Title     *string `gorm:"column:title" json:"title"`
	Company   *string `gorm:"column:company" json:"company"`
	Avatar    *string `gorm:"column:avatar" json:"avatar"`
	Content   string  `gorm:"column:content" json:"content"`
	Rating    int     `gorm:"column:rating" json:"rating"`
	SortOrder int     `gorm:"column:sort_order" json:"sort_order"`
	IsVisible bool    `gorm:"column:is_visible" json:"is_visible"`
	Timestamps
}

func (Testimonial) TableName() string { return "testimonials" }


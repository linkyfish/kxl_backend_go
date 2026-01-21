package model

type Milestone struct {
	ID        int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Year      int    `gorm:"column:year" json:"year"`
	Content   string `gorm:"column:content" json:"content"`
	SortOrder int    `gorm:"column:sort_order" json:"sort_order"`
	Timestamps
}

func (Milestone) TableName() string { return "milestones" }


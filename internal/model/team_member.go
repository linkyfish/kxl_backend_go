package model

type TeamMember struct {
	ID        int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name      string `gorm:"column:name" json:"name"`
	Title     string `gorm:"column:title" json:"title"`
	Avatar    string `gorm:"column:avatar" json:"avatar"`
	Bio       string `gorm:"column:bio" json:"bio"`
	SortOrder int    `gorm:"column:sort_order" json:"sort_order"`
	Timestamps
}

func (TeamMember) TableName() string { return "team_members" }


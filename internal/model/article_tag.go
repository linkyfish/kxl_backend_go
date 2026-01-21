package model

type ArticleTag struct {
	ArticleID string `gorm:"type:uuid;primaryKey;column:article_id" json:"article_id"`
	TagID     int    `gorm:"primaryKey;column:tag_id" json:"tag_id"`
}

func (ArticleTag) TableName() string { return "article_tags" }


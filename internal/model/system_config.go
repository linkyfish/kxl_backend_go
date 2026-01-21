package model

// SystemConfig maps to table `system_configs` (schema is managed by other backends/migrations).
type SystemConfig struct {
	ID          int     `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	GroupName   string  `gorm:"column:group_name" json:"group_name"`
	Key         string  `gorm:"column:key" json:"key"`
	Value       string  `gorm:"column:value" json:"value"`
	Description *string `gorm:"column:description" json:"description"`
	SortOrder   int     `gorm:"column:sort_order" json:"sort_order"`
	IsPublic    bool    `gorm:"column:is_public" json:"is_public"`
	Timestamps
}

func (SystemConfig) TableName() string { return "system_configs" }


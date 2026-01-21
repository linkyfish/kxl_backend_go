package model

type CompanyInfo struct {
	ID            int    `gorm:"primaryKey;autoIncrement;column:id" json:"id"`
	Name          string `gorm:"column:name" json:"name"`
	Description   string `gorm:"column:description" json:"description"`
	Phone         string `gorm:"column:phone" json:"phone"`
	Email         string `gorm:"column:email" json:"email"`
	Address       string `gorm:"column:address" json:"address"`
	WorkingHours  string `gorm:"column:working_hours" json:"working_hours"`
	MapCoordinates string `gorm:"column:map_coordinates" json:"map_coordinates"`
	HeroTitle     string `gorm:"column:hero_title" json:"hero_title"`
	HeroSubtitle  string `gorm:"column:hero_subtitle" json:"hero_subtitle"`

	StatsYears        *string `gorm:"column:stats_years" json:"stats_years"`
	StatsProjects     *string `gorm:"column:stats_projects" json:"stats_projects"`
	StatsClients      *string `gorm:"column:stats_clients" json:"stats_clients"`
	StatsSatisfaction *string `gorm:"column:stats_satisfaction" json:"stats_satisfaction"`

	Timestamps
}

func (CompanyInfo) TableName() string { return "company_info" }

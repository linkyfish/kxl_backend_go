package service

import (
	"context"
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type SettingsService struct {
	db *gorm.DB
}

func NewSettingsService(db *gorm.DB) *SettingsService {
	return &SettingsService{db: db}
}

func (s *SettingsService) ListCategories(ctx context.Context, typ string) ([]model.Category, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.Category{})
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	var rows []model.Category
	if err := q.Order("sort_order asc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SettingsService) CreateCategory(ctx context.Context, name, typ string, sortOrder int) (*model.Category, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	row := &model.Category{Name: name, Type: typ, SortOrder: sortOrder}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return row, nil
}

func (s *SettingsService) UpdateCategory(ctx context.Context, id int, name, typ string, sortOrder int) (*model.Category, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Category
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	row.Name = name
	row.Type = typ
	row.SortOrder = sortOrder
	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SettingsService) DeleteCategory(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}

	var usedProjects, usedArticles, usedCases int64
	_ = s.db.WithContext(ctx).Model(&model.Project{}).Where("category_id = ?", id).Count(&usedProjects).Error
	_ = s.db.WithContext(ctx).Model(&model.Article{}).Where("category_id = ?", id).Count(&usedArticles).Error
	_ = s.db.WithContext(ctx).Model(&model.CaseStudy{}).Where("category_id = ?", id).Count(&usedCases).Error
	if (usedProjects + usedArticles + usedCases) > 0 {
		return kxlerrors.New(kxlerrors.CodeConflict, "conflict: category in use", http.StatusConflict, nil)
	}

	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Category{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: category not found")
	}
	return nil
}

func (s *SettingsService) ListTags(ctx context.Context, typ string) ([]model.Tag, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.Tag{})
	if typ != "" {
		q = q.Where("type = ?", typ)
	}
	var rows []model.Tag
	if err := q.Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SettingsService) CreateTag(ctx context.Context, name, typ string) (*model.Tag, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	row := &model.Tag{Name: name, Type: typ}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return row, nil
}

func (s *SettingsService) UpdateTag(ctx context.Context, id int, name, typ string) (*model.Tag, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Tag
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	row.Name = name
	row.Type = typ
	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SettingsService) DeleteTag(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Tag{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: tag not found")
	}
	return nil
}

func (s *SettingsService) GetCompanyInfo(ctx context.Context) (*model.CompanyInfo, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var info model.CompanyInfo
	err := s.db.WithContext(ctx).Order("id asc").First(&info).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &info, nil
}

func (s *SettingsService) UpsertCompanyInfo(ctx context.Context, payload *model.CompanyInfo) (*model.CompanyInfo, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}

	// Single-row table behavior (match PHP): update first row or create.
	var info model.CompanyInfo
	err := s.db.WithContext(ctx).Order("id asc").First(&info).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, kxlerrors.Internal("db error")
	}

	if err == gorm.ErrRecordNotFound {
		info = model.CompanyInfo{}
	}

	// Copy fields.
	info.Name = payload.Name
	info.Description = payload.Description
	info.Phone = payload.Phone
	info.Email = payload.Email
	info.Address = payload.Address
	info.WorkingHours = payload.WorkingHours
	info.MapCoordinates = payload.MapCoordinates
	info.HeroTitle = payload.HeroTitle
	info.HeroSubtitle = payload.HeroSubtitle

	if err := s.db.WithContext(ctx).Save(&info).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &info, nil
}

func (s *SettingsService) ListMilestones(ctx context.Context) ([]model.Milestone, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Milestone
	if err := s.db.WithContext(ctx).
		Order("year desc").Order("sort_order asc").Order("id asc").
		Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SettingsService) CreateMilestone(ctx context.Context, year int, content string, sortOrder int) (*model.Milestone, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	row := &model.Milestone{Year: year, Content: content, SortOrder: sortOrder}
	if err := s.db.WithContext(ctx).Create(row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return row, nil
}

func (s *SettingsService) UpdateMilestone(ctx context.Context, id int, year int, content string, sortOrder int) (*model.Milestone, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Milestone
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	row.Year = year
	row.Content = content
	row.SortOrder = sortOrder
	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SettingsService) DeleteMilestone(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Milestone{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: milestone not found")
	}
	return nil
}

func (s *SettingsService) ListTeamMembers(ctx context.Context) ([]model.TeamMember, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.TeamMember
	if err := s.db.WithContext(ctx).Order("sort_order asc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SettingsService) CreateTeamMember(ctx context.Context, payload *model.TeamMember) (*model.TeamMember, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *SettingsService) UpdateTeamMember(ctx context.Context, id int, payload *model.TeamMember) (*model.TeamMember, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.TeamMember
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	row.Name = payload.Name
	row.Title = payload.Title
	row.Avatar = payload.Avatar
	row.Bio = payload.Bio
	row.SortOrder = payload.SortOrder
	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SettingsService) DeleteTeamMember(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.TeamMember{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: team member not found")
	}
	return nil
}

func (s *SettingsService) DashboardStats(ctx context.Context) (map[string]int64, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var projects, articles, cases, users, unreadMessages int64
	_ = s.db.WithContext(ctx).Model(&model.Project{}).Count(&projects).Error
	_ = s.db.WithContext(ctx).Model(&model.Article{}).Count(&articles).Error
	_ = s.db.WithContext(ctx).Model(&model.CaseStudy{}).Count(&cases).Error
	_ = s.db.WithContext(ctx).Model(&model.User{}).Count(&users).Error
	_ = s.db.WithContext(ctx).Model(&model.Message{}).Where("status = ?", 0).Count(&unreadMessages).Error

	return map[string]int64{
		"projects":        projects,
		"articles":        articles,
		"cases":           cases,
		"users":           users,
		"unread_messages": unreadMessages,
	}, nil
}


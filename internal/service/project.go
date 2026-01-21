package service

import (
	"context"
	"errors"
	"time"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type ProjectService struct {
	db *gorm.DB
}

func NewProjectService(db *gorm.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) ListPublic(ctx context.Context, page, pageSize int64, categoryID *int, keyword string) ([]model.Project, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.Project{}).Where("status = ?", 1)
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(name ILIKE ? OR description ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}

	var rows []model.Project
	if err := q.Order("sort_order asc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *ProjectService) GetPublic(ctx context.Context, id string) (*model.Project, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var p model.Project
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, kxlerrors.NotFound("not found: project not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	if p.Status != 1 {
		return nil, kxlerrors.NotFound("not found: project not found")
	}
	return &p, nil
}

func (s *ProjectService) ListFeatures(ctx context.Context, projectID string) ([]model.ProjectFeature, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.ProjectFeature
	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("sort_order asc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *ProjectService) ListMedia(ctx context.Context, projectID string) ([]model.ProjectMedia, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.ProjectMedia
	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("sort_order asc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *ProjectService) ListVersions(ctx context.Context, projectID string) ([]model.ProjectVersion, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.ProjectVersion
	if err := s.db.WithContext(ctx).Where("project_id = ?", projectID).Order("release_date desc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *ProjectService) LoadCategoryMap(ctx context.Context, categoryIDs []int) (map[int]model.Category, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if len(categoryIDs) == 0 {
		return map[int]model.Category{}, nil
	}
	var cats []model.Category
	if err := s.db.WithContext(ctx).Where("id in ?", categoryIDs).Find(&cats).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	m := make(map[int]model.Category, len(cats))
	for _, c := range cats {
		m[c.ID] = c
	}
	return m, nil
}

func (s *ProjectService) LoadTagsByProjectIDs(ctx context.Context, projectIDs []string) (map[string][]model.Tag, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if len(projectIDs) == 0 {
		return map[string][]model.Tag{}, nil
	}

	type row struct {
		ProjectID string    `gorm:"column:project_id"`
		ID        int       `gorm:"column:id"`
		Name      string    `gorm:"column:name"`
		Type      string    `gorm:"column:type"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("project_tags").
		Select("project_tags.project_id as project_id, tags.id as id, tags.name as name, tags.type as type, tags.created_at as created_at").
		Joins("join tags on project_tags.tag_id = tags.id").
		Where("project_tags.project_id in ?", projectIDs).
		Order("project_tags.project_id asc").
		Order("tags.id asc").
		Scan(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}

	out := make(map[string][]model.Tag)
	for _, r := range rows {
		out[r.ProjectID] = append(out[r.ProjectID], model.Tag{
			ID:        r.ID,
			Name:      r.Name,
			Type:      r.Type,
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}


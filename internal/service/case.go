package service

import (
	"context"
	"errors"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type CaseService struct {
	db *gorm.DB
}

func NewCaseService(db *gorm.DB) *CaseService {
	return &CaseService{db: db}
}

func (s *CaseService) ListPublic(ctx context.Context, page, pageSize int64, categoryID *int, keyword string) ([]model.CaseStudy, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.CaseStudy{}).Where("status = ?", 1)
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(client_name ILIKE ? OR summary ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}

	var rows []model.CaseStudy
	if err := q.Order("created_at desc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *CaseService) GetPublic(ctx context.Context, id string) (*model.CaseStudy, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var c model.CaseStudy
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&c).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, kxlerrors.NotFound("not found: case not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	if c.Status != 1 {
		return nil, kxlerrors.NotFound("not found: case not found")
	}
	return &c, nil
}

func (s *CaseService) LoadCategoryMap(ctx context.Context, categoryIDs []int) (map[int]model.Category, error) {
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

func (s *CaseService) LoadProjectIDsByCaseID(ctx context.Context, caseID string) ([]string, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	type row struct {
		ProjectID string `gorm:"column:project_id"`
	}
	var rows []row
	if err := s.db.WithContext(ctx).Table("case_projects").Select("project_id").Where("case_id = ?", caseID).Scan(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	out := make([]string, 0, len(rows))
	seen := make(map[string]struct{})
	for _, r := range rows {
		if r.ProjectID == "" {
			continue
		}
		if _, ok := seen[r.ProjectID]; ok {
			continue
		}
		seen[r.ProjectID] = struct{}{}
		out = append(out, r.ProjectID)
	}
	return out, nil
}


package service

import (
	"context"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type SolutionService struct {
	db *gorm.DB
}

func NewSolutionService(db *gorm.DB) *SolutionService {
	return &SolutionService{db: db}
}

func (s *SolutionService) ListVisible(ctx context.Context) ([]model.Solution, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Solution
	if err := s.db.WithContext(ctx).Where("is_visible = ?", true).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SolutionService) ListAll(ctx context.Context) ([]model.Solution, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Solution
	if err := s.db.WithContext(ctx).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SolutionService) Get(ctx context.Context, id int) (*model.Solution, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Solution
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SolutionService) Create(ctx context.Context, payload *model.Solution) (*model.Solution, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *SolutionService) Update(ctx context.Context, id int, payload *model.Solution) (*model.Solution, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Solution
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}

	row.Name = payload.Name
	row.Description = payload.Description
	row.Icon = payload.Icon
	row.BgClass = payload.BgClass
	row.Link = payload.Link
	row.SortOrder = payload.SortOrder
	row.IsVisible = payload.IsVisible

	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SolutionService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	// Match PHP behavior: delete is idempotent.
	_ = s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Solution{}).Error
	return nil
}


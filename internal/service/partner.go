package service

import (
	"context"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type PartnerService struct {
	db *gorm.DB
}

func NewPartnerService(db *gorm.DB) *PartnerService {
	return &PartnerService{db: db}
}

func (s *PartnerService) ListVisible(ctx context.Context) ([]model.Partner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Partner
	if err := s.db.WithContext(ctx).Where("is_visible = ?", true).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *PartnerService) ListAll(ctx context.Context) ([]model.Partner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Partner
	if err := s.db.WithContext(ctx).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *PartnerService) Get(ctx context.Context, id int) (*model.Partner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Partner
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *PartnerService) Create(ctx context.Context, payload *model.Partner) (*model.Partner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *PartnerService) Update(ctx context.Context, id int, payload *model.Partner) (*model.Partner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Partner
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}

	row.Name = payload.Name
	row.Logo = payload.Logo
	row.Website = payload.Website
	row.SortOrder = payload.SortOrder
	row.IsVisible = payload.IsVisible

	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *PartnerService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	// Match PHP behavior: delete is idempotent.
	_ = s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Partner{}).Error
	return nil
}


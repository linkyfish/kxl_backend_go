package service

import (
	"context"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type FriendlyLinkService struct {
	db *gorm.DB
}

func NewFriendlyLinkService(db *gorm.DB) *FriendlyLinkService {
	return &FriendlyLinkService{db: db}
}

func (s *FriendlyLinkService) ListVisible(ctx context.Context) ([]model.FriendlyLink, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.FriendlyLink
	if err := s.db.WithContext(ctx).
		Where("is_visible = ?", true).
		Order("sort_order asc").Order("id asc").
		Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *FriendlyLinkService) ListAll(ctx context.Context) ([]model.FriendlyLink, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.FriendlyLink
	if err := s.db.WithContext(ctx).
		Order("sort_order asc").Order("id asc").
		Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *FriendlyLinkService) Get(ctx context.Context, id int) (*model.FriendlyLink, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.FriendlyLink
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *FriendlyLinkService) Create(ctx context.Context, payload *model.FriendlyLink) (*model.FriendlyLink, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *FriendlyLinkService) Update(ctx context.Context, id int, payload *model.FriendlyLink) (*model.FriendlyLink, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.FriendlyLink
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}

	row.Name = payload.Name
	row.URL = payload.URL
	row.Logo = payload.Logo
	row.Description = payload.Description
	row.SortOrder = payload.SortOrder
	row.IsVisible = payload.IsVisible

	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *FriendlyLinkService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.FriendlyLink{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: friendly link not found")
	}
	return nil
}

func (s *FriendlyLinkService) BatchDelete(ctx context.Context, ids []int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	if len(ids) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Where("id in ?", ids).Delete(&model.FriendlyLink{}).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return nil
}


package service

import (
	"context"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type BannerService struct {
	db *gorm.DB
}

func NewBannerService(db *gorm.DB) *BannerService {
	return &BannerService{db: db}
}

func (s *BannerService) ListVisible(ctx context.Context) ([]model.Banner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Banner
	if err := s.db.WithContext(ctx).Where("is_visible = ?", true).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *BannerService) ListAll(ctx context.Context) ([]model.Banner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Banner
	if err := s.db.WithContext(ctx).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *BannerService) Get(ctx context.Context, id int) (*model.Banner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Banner
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *BannerService) Create(ctx context.Context, payload *model.Banner) (*model.Banner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *BannerService) Update(ctx context.Context, id int, payload *model.Banner) (*model.Banner, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Banner
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}

	row.Title = payload.Title
	row.Subtitle = payload.Subtitle
	row.Highlight = payload.Highlight
	row.Tag = payload.Tag
	row.Image = payload.Image
	row.Link = payload.Link
	row.LinkText = payload.LinkText
	row.BgClass = payload.BgClass
	row.SortOrder = payload.SortOrder
	row.IsVisible = payload.IsVisible

	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *BannerService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	// Match PHP behavior: delete is idempotent.
	_ = s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Banner{}).Error
	return nil
}


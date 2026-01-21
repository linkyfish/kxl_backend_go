package service

import (
	"context"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type TestimonialService struct {
	db *gorm.DB
}

func NewTestimonialService(db *gorm.DB) *TestimonialService {
	return &TestimonialService{db: db}
}

func (s *TestimonialService) ListVisible(ctx context.Context) ([]model.Testimonial, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Testimonial
	if err := s.db.WithContext(ctx).Where("is_visible = ?", true).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *TestimonialService) ListAll(ctx context.Context) ([]model.Testimonial, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Testimonial
	if err := s.db.WithContext(ctx).Order("sort_order asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *TestimonialService) Get(ctx context.Context, id int) (*model.Testimonial, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Testimonial
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *TestimonialService) Create(ctx context.Context, payload *model.Testimonial) (*model.Testimonial, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *TestimonialService) Update(ctx context.Context, id int, payload *model.Testimonial) (*model.Testimonial, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.Testimonial
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}

	row.Name = payload.Name
	row.Title = payload.Title
	row.Company = payload.Company
	row.Avatar = payload.Avatar
	row.Content = payload.Content
	row.Rating = payload.Rating
	row.SortOrder = payload.SortOrder
	row.IsVisible = payload.IsVisible

	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *TestimonialService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	// Match PHP behavior: delete is idempotent.
	_ = s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Testimonial{}).Error
	return nil
}


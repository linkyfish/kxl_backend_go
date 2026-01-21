package service

import (
	"context"
	"errors"
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type SystemConfigService struct {
	db *gorm.DB
}

func NewSystemConfigService(db *gorm.DB) *SystemConfigService {
	return &SystemConfigService{db: db}
}

func (s *SystemConfigService) List(ctx context.Context, group string) ([]model.SystemConfig, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.SystemConfig{})
	if group != "" {
		q = q.Where("group_name = ?", group)
	}
	var rows []model.SystemConfig
	if err := q.Order("group_name asc").Order("sort_order asc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *SystemConfigService) Create(ctx context.Context, payload *model.SystemConfig) (*model.SystemConfig, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if err := s.db.WithContext(ctx).Create(payload).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, kxlerrors.New(kxlerrors.CodeConflict, "conflict: 配置项已存在（分组 + key 必须唯一）", http.StatusConflict, nil)
		}
		return nil, kxlerrors.Internal("db error")
	}
	return payload, nil
}

func (s *SystemConfigService) Update(ctx context.Context, id int, payload *model.SystemConfig) (*model.SystemConfig, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var row model.SystemConfig
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}

	row.GroupName = payload.GroupName
	row.Key = payload.Key
	row.Value = payload.Value
	row.Description = payload.Description
	row.SortOrder = payload.SortOrder
	row.IsPublic = payload.IsPublic

	if err := s.db.WithContext(ctx).Save(&row).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, kxlerrors.New(kxlerrors.CodeConflict, "conflict: 配置项已存在（分组 + key 必须唯一）", http.StatusConflict, nil)
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &row, nil
}

func (s *SystemConfigService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.SystemConfig{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: system config not found")
	}
	return nil
}


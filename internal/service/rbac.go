package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type RbacService struct {
	db    *gorm.DB
	redis *redis.Client
	ttl   time.Duration
}

func NewRbacService(db *gorm.DB, redisClient *redis.Client, ttlSeconds int) *RbacService {
	ttl := time.Duration(ttlSeconds) * time.Second
	if ttlSeconds <= 0 {
		ttl = 300 * time.Second
	}
	return &RbacService{
		db:    db,
		redis: redisClient,
		ttl:   ttl,
	}
}

func (s *RbacService) GetPermissionsForRole(ctx context.Context, roleCode string) ([]string, error) {
	if roleCode == "super_admin" {
		return []string{"*"}, nil
	}
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}

	var rows []model.AdminRolePermission
	if err := s.db.WithContext(ctx).
		Where("role_code = ?", roleCode).
		Order("permission_code asc").
		Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	perms := make([]string, 0, len(rows))
	for _, r := range rows {
		perms = append(perms, r.PermissionCode)
	}
	return perms, nil
}

func (s *RbacService) GetCachedPermissionsForRole(ctx context.Context, roleCode string) ([]string, error) {
	if roleCode == "super_admin" {
		return []string{"*"}, nil
	}
	if s == nil {
		return nil, kxlerrors.Internal("rbac not configured")
	}

	cacheKey := fmt.Sprintf("rbac:role_permissions:%s", roleCode)
	if s.redis != nil {
		if raw, err := s.redis.Get(ctx, cacheKey).Bytes(); err == nil && len(raw) > 0 {
			var decoded []string
			if json.Unmarshal(raw, &decoded) == nil {
				return decoded, nil
			}
		}
	}

	perms, err := s.GetPermissionsForRole(ctx, roleCode)
	if err != nil {
		return nil, err
	}

	if s.redis != nil && s.ttl > 0 {
		payload, _ := json.Marshal(perms)
		_ = s.redis.SetEX(ctx, cacheKey, payload, s.ttl).Err()
	}
	return perms, nil
}

func (s *RbacService) InvalidateRolePermissions(ctx context.Context, roleCode string) {
	if roleCode == "super_admin" || s == nil || s.redis == nil {
		return
	}
	cacheKey := fmt.Sprintf("rbac:role_permissions:%s", roleCode)
	_ = s.redis.Del(ctx, cacheKey).Err()
}

func HasPermission(role string, permissions []string, code string) bool {
	if role == "super_admin" {
		return true
	}
	for _, p := range permissions {
		if p == "*" || p == code {
			return true
		}
	}
	return false
}

func SortPermissions(perms []string) []string {
	out := append([]string{}, perms...)
	sort.Strings(out)
	return out
}


package service

import (
	"context"
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/util"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

func (s *UserService) ListUsers(ctx context.Context, page, pageSize int64, keyword string, status *int) ([]model.User, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, kxlerrors.Internal("db not configured")
	}

	q := s.db.WithContext(ctx).Model(&model.User{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(username ILIKE ? OR email ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}

	var rows []model.User
	if err := q.Order("created_at desc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*model.User, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &user, nil
}

func (s *UserService) UpdateUserStatus(ctx context.Context, id string, status int16) (*model.User, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var user model.User
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	if user.Status == status {
		return &user, nil
	}
	user.Status = status
	if status == 0 {
		user.SessionVersion = user.SessionVersion + 1
	}
	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &user, nil
}

func (s *UserService) ListAdmins(ctx context.Context) ([]model.Admin, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var rows []model.Admin
	if err := s.db.WithContext(ctx).Order("created_at desc").Order("id asc").Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *UserService) CreateAdmin(ctx context.Context, username, password, role string, status int16) (*model.Admin, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&model.Admin{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	if count > 0 {
		return nil, kxlerrors.New(kxlerrors.CodeConflict, "conflict: username already exists", http.StatusConflict, nil)
	}

	hashed, err := util.HashPassword(password)
	if err != nil {
		return nil, kxlerrors.Internal("password hash error")
	}

	admin := &model.Admin{
		Username:     username,
		PasswordHash: hashed,
		Role:         role,
		Status:       status,
	}
	if err := s.db.WithContext(ctx).Create(admin).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return admin, nil
}

func (s *UserService) UpdateAdmin(ctx context.Context, id, username, role string, status int16) (*model.Admin, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var admin model.Admin
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	admin.Username = username
	admin.Role = role
	admin.Status = status
	if err := s.db.WithContext(ctx).Save(&admin).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &admin, nil
}

func (s *UserService) DeleteAdmin(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	var admin model.Admin
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&admin).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}

	if admin.Role == "super_admin" {
		var count int64
		_ = s.db.WithContext(ctx).Model(&model.Admin{}).Where("role = ?", "super_admin").Count(&count).Error
		if count <= 1 {
			return kxlerrors.New(kxlerrors.CodeConflict, "conflict: cannot delete last super admin", http.StatusConflict, nil)
		}
	}

	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Admin{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: admin not found")
	}
	return nil
}


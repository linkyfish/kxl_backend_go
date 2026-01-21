package service

import (
	"context"
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/util"
	"gorm.io/gorm"
)

type AuthService struct {
	db *gorm.DB
}

func NewAuthService(db *gorm.DB) *AuthService {
	return &AuthService{db: db}
}

func (s *AuthService) RegisterUser(ctx context.Context, username, email, password string) (*model.User, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}

	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("username = ?", username).Count(&count).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	if count > 0 {
		return nil, kxlerrors.New(kxlerrors.CodeConflict, "conflict: username already exists", http.StatusConflict, nil)
	}
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	if count > 0 {
		return nil, kxlerrors.New(kxlerrors.CodeConflict, "conflict: email already exists", http.StatusConflict, nil)
	}

	hashed, err := util.HashPassword(password)
	if err != nil {
		return nil, kxlerrors.Internal("password hash error")
	}

	user := &model.User{
		Username:       username,
		Email:          email,
		PasswordHash:   hashed,
		Status:         1,
		SessionVersion: 0,
	}
	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return user, nil
}

func (s *AuthService) AuthenticateUser(ctx context.Context, identifier, password string) (*model.User, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}

	var user model.User
	err := s.db.WithContext(ctx).
		Where("username = ? OR email = ?", identifier, identifier).
		First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	if user.Status != 1 {
		return nil, kxlerrors.Forbidden()
	}
	if !util.CheckPassword(user.PasswordHash, password) {
		return nil, kxlerrors.Unauthorized()
	}
	return &user, nil
}

func (s *AuthService) AuthenticateAdmin(ctx context.Context, username, password string) (*model.Admin, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}

	var admin model.Admin
	err := s.db.WithContext(ctx).
		Where("username = ?", username).
		First(&admin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	if admin.Status != 1 {
		return nil, kxlerrors.Forbidden()
	}
	if !util.CheckPassword(admin.PasswordHash, password) {
		return nil, kxlerrors.Unauthorized()
	}
	return &admin, nil
}


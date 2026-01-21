package middleware

import (
	"errors"

	"github.com/go-redis/redis/v8"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func AuthAdmin(db *gorm.DB, sess *session.Manager, rbac *service.RbacService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if db == nil || sess == nil || sess.Client == nil {
				return kxlerrors.Internal("auth middleware not configured")
			}

			sid, err := c.Cookie(sess.AdminCookieName)
			if err != nil || sid == nil || sid.Value == "" {
				return kxlerrors.Unauthorized()
			}

			ctx := c.Request().Context()
			s, err := sess.GetAdminSession(ctx, sid.Value)
			if err != nil {
				if errors.Is(err, redis.Nil) {
					return kxlerrors.Unauthorized()
				}
				return kxlerrors.Internal("session backend error")
			}
			if s.AdminID == "" {
				_ = sess.DeleteAdminSession(ctx, sid.Value)
				return kxlerrors.Unauthorized()
			}

			var admin model.Admin
			if err := db.Where("id = ?", s.AdminID).First(&admin).Error; err != nil {
				_ = sess.DeleteAdminSession(ctx, sid.Value)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return kxlerrors.Unauthorized()
				}
				return kxlerrors.Internal("db error")
			}
			if admin.Status != 1 {
				_ = sess.DeleteAdminSession(ctx, sid.Value)
				return kxlerrors.Unauthorized()
			}

			permissions := []string{}
			if admin.Role == "super_admin" {
				permissions = []string{"*"}
			} else if rbac != nil {
				if perms, err := rbac.GetCachedPermissionsForRole(ctx, admin.Role); err == nil {
					permissions = perms
				}
			}

			c.Set("current_admin_id", admin.ID)
			c.Set("current_admin_role", admin.Role)
			c.Set("current_admin_permissions", permissions)
			c.Set("current_admin", &admin)
			return next(c)
		}
	}
}

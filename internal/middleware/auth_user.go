package middleware

import (
	"errors"

	"github.com/go-redis/redis/v8"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

func AuthUser(db *gorm.DB, sess *session.Manager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if db == nil || sess == nil || sess.Client == nil {
				return kxlerrors.Internal("auth middleware not configured")
			}

			sid, err := c.Cookie(sess.UserCookieName)
			if err != nil || sid == nil || sid.Value == "" {
				return kxlerrors.Unauthorized()
			}

			ctx := c.Request().Context()
			s, err := sess.GetUserSession(ctx, sid.Value)
			if err != nil {
				if errors.Is(err, redis.Nil) {
					return kxlerrors.Unauthorized()
				}
				return kxlerrors.Internal("session backend error")
			}
			if s.UserID == "" {
				_ = sess.DeleteUserSession(ctx, sid.Value)
				return kxlerrors.Unauthorized()
			}

			var user model.User
			if err := db.Where("id = ?", s.UserID).First(&user).Error; err != nil {
				_ = sess.DeleteUserSession(ctx, sid.Value)
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return kxlerrors.Unauthorized()
				}
				return kxlerrors.Internal("db error")
			}

			if user.Status != 1 || user.SessionVersion != s.UserSessionVersion {
				_ = sess.DeleteUserSession(ctx, sid.Value)
				return kxlerrors.Unauthorized()
			}

			c.Set("current_user_id", user.ID)
			c.Set("current_user", &user)
			return next(c)
		}
	}
}

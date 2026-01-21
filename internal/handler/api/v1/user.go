package v1

import (
	"net/http"
	"time"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/util"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type UserHandler struct {
	DB       *gorm.DB
	Sessions *session.Manager
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" form:"old_password"`
	NewPassword string `json:"new_password" form:"new_password"`
}

func (h *UserHandler) ChangePassword(c echo.Context) error {
	u, _ := c.Get("current_user").(*model.User)
	if u == nil {
		return kxlerrors.Unauthorized()
	}

	var req changePasswordRequest
	_ = c.Bind(&req)
	if req.OldPassword == "" || req.NewPassword == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	if !util.CheckPassword(u.PasswordHash, req.OldPassword) {
		return kxlerrors.Unauthorized()
	}

	hashed, err := util.HashPassword(req.NewPassword)
	if err != nil {
		return kxlerrors.Internal("password hash error")
	}

	u.PasswordHash = hashed
	u.SessionVersion = u.SessionVersion + 1
	if err := h.DB.WithContext(c.Request().Context()).Save(u).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	// Purge current session (force re-login), matching Rust/PHP behavior.
	if h.Sessions != nil {
		if cookie, err := c.Cookie(h.Sessions.UserCookieName); err == nil && cookie != nil && cookie.Value != "" {
			_ = h.Sessions.DeleteUserSession(c.Request().Context(), cookie.Value)
		}
		h.clearCookie(c, h.Sessions.UserCookieName)
	}

	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *UserHandler) clearCookie(c echo.Context, name string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Sessions != nil && h.Sessions.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
	}
	c.SetCookie(cookie)
}


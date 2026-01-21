package admin

import (
	"net/http"
	"time"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Auth     *service.AuthService
	RBAC     *service.RbacService
	Sessions *session.Manager
}

type loginRequest struct {
	Identifier string `json:"identifier" form:"identifier"`
	Password   string `json:"password" form:"password"`
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	_ = c.Bind(&req)
	if req.Identifier == "" || req.Password == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	admin, err := h.Auth.AuthenticateAdmin(c.Request().Context(), req.Identifier, req.Password)
	if err != nil {
		return err
	}

	sid, err := h.Sessions.CreateAdminSession(c.Request().Context(), admin.ID)
	if err != nil {
		return kxlerrors.Internal("session backend error")
	}

	perms, _ := h.RBAC.GetPermissionsForRole(c.Request().Context(), admin.Role)
	h.setCookie(c, h.Sessions.AdminCookieName, sid, h.Sessions.AdminTTL)

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          admin.ID,
		"username":    admin.Username,
		"role":        admin.Role,
		"permissions": perms,
		"status":      admin.Status,
		"created_at":  admin.CreatedAt,
	}))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	if h.Sessions != nil {
		if cookie, err := c.Cookie(h.Sessions.AdminCookieName); err == nil && cookie != nil && cookie.Value != "" {
			_ = h.Sessions.DeleteAdminSession(c.Request().Context(), cookie.Value)
		}
		h.clearCookie(c, h.Sessions.AdminCookieName)
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *AuthHandler) Me(c echo.Context) error {
	admin, _ := c.Get("current_admin").(*model.Admin)
	if admin == nil {
		return kxlerrors.Unauthorized()
	}

	perms, _ := h.RBAC.GetPermissionsForRole(c.Request().Context(), admin.Role)
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          admin.ID,
		"username":    admin.Username,
		"role":        admin.Role,
		"permissions": perms,
		"status":      admin.Status,
		"created_at":  admin.CreatedAt,
	}))
}

func (h *AuthHandler) setCookie(c echo.Context, name, value string, ttl time.Duration) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.Sessions != nil && h.Sessions.CookieSecure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
	}
	c.SetCookie(cookie)
}

func (h *AuthHandler) clearCookie(c echo.Context, name string) {
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


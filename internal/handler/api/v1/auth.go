package v1

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
	Sessions *session.Manager
}

type registerRequest struct {
	Username string `json:"username" form:"username"`
	Email    string `json:"email" form:"email"`
	Password string `json:"password" form:"password"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	_ = c.Bind(&req)
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	user, err := h.Auth.RegisterUser(c.Request().Context(), req.Username, req.Email, req.Password)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	}))
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

	user, err := h.Auth.AuthenticateUser(c.Request().Context(), req.Identifier, req.Password)
	if err != nil {
		return err
	}

	sid, err := h.Sessions.CreateUserSession(c.Request().Context(), user.ID, user.SessionVersion)
	if err != nil {
		return kxlerrors.Internal("session backend error")
	}

	h.setCookie(c, h.Sessions.UserCookieName, sid, h.Sessions.UserTTL)

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         user.ID,
		"username":   user.Username,
		"email":      user.Email,
		"status":     user.Status,
		"created_at": user.CreatedAt,
	}))
}

func (h *AuthHandler) Logout(c echo.Context) error {
	if h.Sessions != nil {
		if cookie, err := c.Cookie(h.Sessions.UserCookieName); err == nil && cookie != nil && cookie.Value != "" {
			_ = h.Sessions.DeleteUserSession(c.Request().Context(), cookie.Value)
		}
		h.clearCookie(c, h.Sessions.UserCookieName)
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *AuthHandler) Me(c echo.Context) error {
	u, _ := c.Get("current_user").(*model.User)
	if u == nil {
		return kxlerrors.Unauthorized()
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"status":     u.Status,
		"created_at": u.CreatedAt,
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


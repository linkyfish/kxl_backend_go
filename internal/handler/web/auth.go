package web

import (
	"net/http"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/linkyfish/kxl_backend_go/pkg/session"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
	Auth     *service.AuthService
	Sessions *session.Manager
}

func (h *AuthHandler) LoginPage(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}
	ctx := pongo2.Context{
		"page_title": "登录",
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/login.html", ctx)
}

func (h *AuthHandler) LoginSubmit(c echo.Context) error {
	identifier := strings.TrimSpace(c.FormValue("username"))
	if identifier == "" {
		identifier = strings.TrimSpace(c.FormValue("identifier"))
	}
	password := c.FormValue("password")

	if identifier == "" || password == "" {
		return h.loginError(c, identifier, "请输入用户名和密码")
	}

	user, err := h.Auth.AuthenticateUser(c.Request().Context(), identifier, password)
	if err != nil {
		if be, ok := err.(*kxlerrors.BusinessError); ok {
			if be.HTTPStatus == http.StatusUnauthorized || be.HTTPStatus == http.StatusForbidden || be.HTTPStatus == http.StatusNotFound {
				return h.loginError(c, identifier, "用户名或密码错误")
			}
		}
		return err
	}

	sid, err := h.Sessions.CreateUserSession(c.Request().Context(), user.ID, user.SessionVersion)
	if err != nil {
		return err
	}
	setCookie(c, h.Sessions.UserCookieName, sid, h.Sessions.UserTTL, h.Sessions.CookieSecure)

	if wantsJSON(c.Request()) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     200,
			"message":  "success",
			"redirect": "/",
		})
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) RegisterPage(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}
	ctx := pongo2.Context{
		"page_title": "注册",
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/register.html", ctx)
}

func (h *AuthHandler) RegisterSubmit(c echo.Context) error {
	username := strings.TrimSpace(c.FormValue("username"))
	email := strings.TrimSpace(c.FormValue("email"))
	password := c.FormValue("password")
	confirm := c.FormValue("password_confirm")

	if username == "" || email == "" || password == "" || confirm == "" {
		return h.registerError(c, username, email, "请填写必填字段")
	}
	if password != confirm {
		return h.registerError(c, username, email, "两次输入的密码不一致")
	}

	user, err := h.Auth.RegisterUser(c.Request().Context(), username, email, password)
	if err != nil {
		// Pass through business error message (conflict, validation, etc.).
		return h.registerError(c, username, email, err.Error())
	}

	sid, err := h.Sessions.CreateUserSession(c.Request().Context(), user.ID, user.SessionVersion)
	if err != nil {
		return err
	}
	setCookie(c, h.Sessions.UserCookieName, sid, h.Sessions.UserTTL, h.Sessions.CookieSecure)

	if wantsJSON(c.Request()) {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"code":     200,
			"message":  "success",
			"redirect": "/",
		})
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) Logout(c echo.Context) error {
	if h.Sessions != nil {
		if cookie, err := c.Cookie(h.Sessions.UserCookieName); err == nil && cookie != nil && cookie.Value != "" {
			_ = h.Sessions.DeleteUserSession(c.Request().Context(), cookie.Value)
		}
		clearCookie(c, h.Sessions.UserCookieName, h.Sessions.CookieSecure)
	}
	return c.Redirect(http.StatusSeeOther, "/")
}

func (h *AuthHandler) loginError(c echo.Context, username, msg string) error {
	if wantsJSON(c.Request()) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": msg,
		})
	}

	base, _ := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	ctx := pongo2.Context{
		"page_title": "登录",
		"error":      msg,
		"username":   username,
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/login.html", ctx)
}

func (h *AuthHandler) registerError(c echo.Context, username, email, msg string) error {
	if wantsJSON(c.Request()) {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": msg,
		})
	}

	base, _ := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	ctx := pongo2.Context{
		"page_title": "注册",
		"error":      msg,
		"username":   username,
		"email":      email,
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/register.html", ctx)
}

func setCookie(c echo.Context, name, value string, ttl time.Duration, secure bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
		Expires:  time.Now().Add(ttl),
	}
	c.SetCookie(cookie)
}

func clearCookie(c echo.Context, name string, secure bool) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   0,
		Expires:  time.Unix(0, 0),
	}
	c.SetCookie(cookie)
}

package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

// ErrorHandler renders common error pages. The global HTTPErrorHandler uses the same templates,
// but having a dedicated handler keeps SSR behavior explicit and testable.
type ErrorHandler struct {
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
}

func (h *ErrorHandler) NotFound(c echo.Context) error {
	base, _ := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	ctx := pongo2.Context{
		"page_title": "404 - 页面未找到",
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusNotFound, "pages/error/404.html", ctx)
}

func (h *ErrorHandler) Internal(c echo.Context) error {
	base, _ := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	ctx := pongo2.Context{
		"page_title": "500 - 服务器错误",
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusInternalServerError, "pages/error/500.html", ctx)
}


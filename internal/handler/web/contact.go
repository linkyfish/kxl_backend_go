package web

import (
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type ContactHandler struct {
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
	Messages *service.MessageService
}

func (h *ContactHandler) Index(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	ctx := pongo2.Context{
		"page_title":  "联系我们",
		"breadcrumbs": []map[string]interface{}{{"title": "联系我们", "url": "/contact"}},
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/contact.html", ctx)
}

func (h *ContactHandler) Submit(c echo.Context) error {
	// This endpoint is used by the SSR contact form (data-form) and expects JSON.
	name := strings.TrimSpace(c.FormValue("name"))
	email := strings.TrimSpace(c.FormValue("email"))
	phone := strings.TrimSpace(c.FormValue("phone"))
	company := strings.TrimSpace(c.FormValue("company"))
	content := strings.TrimSpace(c.FormValue("content"))
	subject := strings.TrimSpace(c.FormValue("subject"))

	if subject != "" {
		// Keep the subject for operators; match existing backends where subject is part of content.
		content = "[" + subject + "] " + content
	}

	if name == "" || email == "" || content == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"code":    400,
			"message": "请填写必填字段",
		})
	}

	var companyPtr *string
	if company != "" {
		companyPtr = &company
	}

	if h.Messages != nil {
		if _, err := h.Messages.Submit(c.Request().Context(), name, companyPtr, phone, email, content); err != nil {
			return err
		}
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"code":    200,
		"message": "留言提交成功，我们会尽快与您联系！",
	})
}


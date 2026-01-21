package v1

import (
	"net/http"

	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type FriendlyLinkHandler struct {
	FriendlyLinks *service.FriendlyLinkService
}

func (h *FriendlyLinkHandler) List(c echo.Context) error {
	items, err := h.FriendlyLinks.ListVisible(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, f := range items {
		data = append(data, map[string]interface{}{
			"id":          f.ID,
			"name":        f.Name,
			"url":         f.URL,
			"logo":        f.Logo,
			"description": f.Description,
			"sort_order":  f.SortOrder,
			"is_visible":  f.IsVisible,
			"created_at":  f.CreatedAt,
			"updated_at":  f.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}


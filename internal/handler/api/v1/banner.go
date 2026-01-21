package v1

import (
	"net/http"

	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type BannerHandler struct {
	Banners *service.BannerService
}

func (h *BannerHandler) List(c echo.Context) error {
	items, err := h.Banners.ListVisible(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, b := range items {
		data = append(data, map[string]interface{}{
			"id":         b.ID,
			"title":      b.Title,
			"subtitle":   b.Subtitle,
			"highlight":  b.Highlight,
			"tag":        b.Tag,
			"image":      b.Image,
			"link":       b.Link,
			"link_text":  b.LinkText,
			"bg_class":   b.BgClass,
			"sort_order": b.SortOrder,
			"is_visible": b.IsVisible,
			"created_at": b.CreatedAt,
			"updated_at": b.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}


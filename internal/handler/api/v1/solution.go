package v1

import (
	"net/http"

	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type SolutionHandler struct {
	Solutions *service.SolutionService
}

func (h *SolutionHandler) List(c echo.Context) error {
	items, err := h.Solutions.ListVisible(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, s := range items {
		data = append(data, map[string]interface{}{
			"id":         s.ID,
			"name":       s.Name,
			"description": s.Description,
			"icon":       s.Icon,
			"bg_class":   s.BgClass,
			"link":       s.Link,
			"sort_order": s.SortOrder,
			"is_visible": s.IsVisible,
			"created_at": s.CreatedAt,
			"updated_at": s.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}


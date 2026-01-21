package v1

import (
	"net/http"

	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type PartnerHandler struct {
	Partners *service.PartnerService
}

func (h *PartnerHandler) List(c echo.Context) error {
	items, err := h.Partners.ListVisible(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, p := range items {
		data = append(data, map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"logo":       p.Logo,
			"website":    p.Website,
			"sort_order": p.SortOrder,
			"is_visible": p.IsVisible,
			"created_at": p.CreatedAt,
			"updated_at": p.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}


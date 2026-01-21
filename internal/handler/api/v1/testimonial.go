package v1

import (
	"net/http"

	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type TestimonialHandler struct {
	Testimonials *service.TestimonialService
}

func (h *TestimonialHandler) List(c echo.Context) error {
	items, err := h.Testimonials.ListVisible(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, t := range items {
		data = append(data, map[string]interface{}{
			"id":         t.ID,
			"name":       t.Name,
			"title":      t.Title,
			"company":    t.Company,
			"avatar":     t.Avatar,
			"content":    t.Content,
			"rating":     t.Rating,
			"sort_order": t.SortOrder,
			"is_visible": t.IsVisible,
			"created_at": t.CreatedAt,
			"updated_at": t.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}


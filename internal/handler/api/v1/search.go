package v1

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/service"
)

type SearchHandler struct {
	SearchSvc *service.SearchService
}

func (h *SearchHandler) Search(c echo.Context) error {
	q := c.QueryParam("q")
	if q == "" {
		return kxlerrors.Validation("validation error: q is required")
	}
	typ := c.QueryParam("type")
	var typPtr *string
	if typ != "" {
		typPtr = &typ
	}

	page := int64(1)
	pageSize := int64(10)
	if raw := c.QueryParam("page"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			page = n
		}
	}
	if raw := c.QueryParam("page_size"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			pageSize = n
		}
	}
	if page < 1 {
		return kxlerrors.Validation("validation error: page must be >= 1")
	}
	if pageSize < 1 || pageSize > 200 {
		return kxlerrors.Validation("validation error: page_size must be between 1 and 200")
	}

	items, total, err := h.SearchSvc.Search(c.Request().Context(), q, typPtr, page, pageSize)
	if err != nil {
		return err
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}))
}

func (h *SearchHandler) Suggestions(c echo.Context) error {
	items, _ := h.SearchSvc.Suggestions(c.Request().Context())
	return c.JSON(http.StatusOK, response.Success(items))
}

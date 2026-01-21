package admin

import (
	"net/http"
	"strconv"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type FriendlyLinkHandler struct {
	FriendlyLinks *service.FriendlyLinkService
}

func (h *FriendlyLinkHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.FriendlyLinks.ListAll(c.Request().Context())
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

func (h *FriendlyLinkHandler) Detail(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	f, err := h.FriendlyLinks.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          f.ID,
		"name":        f.Name,
		"url":         f.URL,
		"logo":        f.Logo,
		"description": f.Description,
		"sort_order":  f.SortOrder,
		"is_visible":  f.IsVisible,
		"created_at":  f.CreatedAt,
		"updated_at":  f.UpdatedAt,
	}))
}

type friendlyLinkRequest struct {
	Name        string  `json:"name" form:"name"`
	URL         string  `json:"url" form:"url"`
	Logo        *string `json:"logo" form:"logo"`
	Description *string `json:"description" form:"description"`
	SortOrder   int     `json:"sort_order" form:"sort_order"`
	IsVisible   *bool   `json:"is_visible" form:"is_visible"`
}

func (h *FriendlyLinkHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req friendlyLinkRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.URL == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.FriendlyLinks.Create(c.Request().Context(), &model.FriendlyLink{
		Name:        req.Name,
		URL:         req.URL,
		Logo:        normalizeOptString(req.Logo),
		Description: normalizeOptString(req.Description),
		SortOrder:   req.SortOrder,
		IsVisible:   isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          row.ID,
		"name":        row.Name,
		"url":         row.URL,
		"logo":        row.Logo,
		"description": row.Description,
		"sort_order":  row.SortOrder,
		"is_visible":  row.IsVisible,
		"created_at":  row.CreatedAt,
		"updated_at":  row.UpdatedAt,
	}))
}

func (h *FriendlyLinkHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req friendlyLinkRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.URL == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.FriendlyLinks.Update(c.Request().Context(), id, &model.FriendlyLink{
		Name:        req.Name,
		URL:         req.URL,
		Logo:        normalizeOptString(req.Logo),
		Description: normalizeOptString(req.Description),
		SortOrder:   req.SortOrder,
		IsVisible:   isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          row.ID,
		"name":        row.Name,
		"url":         row.URL,
		"logo":        row.Logo,
		"description": row.Description,
		"sort_order":  row.SortOrder,
		"is_visible":  row.IsVisible,
		"created_at":  row.CreatedAt,
		"updated_at":  row.UpdatedAt,
	}))
}

func (h *FriendlyLinkHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.FriendlyLinks.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type batchDeleteIDsRequest struct {
	IDs []int `json:"ids" form:"ids"`
}

func (h *FriendlyLinkHandler) BatchDelete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req batchDeleteIDsRequest
	_ = c.Bind(&req)
	if req.IDs == nil || len(req.IDs) == 0 {
		return c.JSON(http.StatusOK, response.SuccessWithoutData())
	}
	if err := h.FriendlyLinks.BatchDelete(c.Request().Context(), req.IDs); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func normalizeOptString(s *string) *string {
	if s == nil {
		return nil
	}
	if *s == "" {
		return nil
	}
	return s
}


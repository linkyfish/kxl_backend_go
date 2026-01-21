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

type SolutionHandler struct {
	Solutions *service.SolutionService
}

func (h *SolutionHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.Solutions.ListAll(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, s := range items {
		data = append(data, solutionDTO(s))
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

func (h *SolutionHandler) Detail(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	row, err := h.Solutions.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(solutionDTO(*row)))
}

type solutionRequest struct {
	Name        string  `json:"name" form:"name"`
	Description string  `json:"description" form:"description"`
	Icon        *string `json:"icon" form:"icon"`
	BgClass     string  `json:"bg_class" form:"bg_class"`
	Link        string  `json:"link" form:"link"`
	SortOrder   int     `json:"sort_order" form:"sort_order"`
	IsVisible   *bool   `json:"is_visible" form:"is_visible"`
}

func (h *SolutionHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req solutionRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Description == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.Solutions.Create(c.Request().Context(), &model.Solution{
		Name:        req.Name,
		Description: req.Description,
		Icon:        normalizeOptString(req.Icon),
		BgClass:     req.BgClass,
		Link:        req.Link,
		SortOrder:   req.SortOrder,
		IsVisible:   isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(solutionDTO(*row)))
}

func (h *SolutionHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req solutionRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Description == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.Solutions.Update(c.Request().Context(), id, &model.Solution{
		Name:        req.Name,
		Description: req.Description,
		Icon:        normalizeOptString(req.Icon),
		BgClass:     req.BgClass,
		Link:        req.Link,
		SortOrder:   req.SortOrder,
		IsVisible:   isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(solutionDTO(*row)))
}

func (h *SolutionHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	_ = h.Solutions.Delete(c.Request().Context(), id)
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func solutionDTO(s model.Solution) map[string]interface{} {
	return map[string]interface{}{
		"id":          s.ID,
		"name":        s.Name,
		"description": s.Description,
		"icon":        s.Icon,
		"bg_class":    s.BgClass,
		"link":        s.Link,
		"sort_order":  s.SortOrder,
		"is_visible":  s.IsVisible,
		"created_at":  s.CreatedAt,
		"updated_at":  s.UpdatedAt,
	}
}


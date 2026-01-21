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

type PartnerHandler struct {
	Partners *service.PartnerService
}

func (h *PartnerHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.Partners.ListAll(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, p := range items {
		data = append(data, partnerDTO(p))
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

func (h *PartnerHandler) Detail(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	row, err := h.Partners.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(partnerDTO(*row)))
}

type partnerRequest struct {
	Name      string  `json:"name" form:"name"`
	Logo      *string `json:"logo" form:"logo"`
	Website   *string `json:"website" form:"website"`
	SortOrder int     `json:"sort_order" form:"sort_order"`
	IsVisible *bool   `json:"is_visible" form:"is_visible"`
}

func (h *PartnerHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req partnerRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.Partners.Create(c.Request().Context(), &model.Partner{
		Name:      req.Name,
		Logo:      normalizeOptString(req.Logo),
		Website:   normalizeOptString(req.Website),
		SortOrder: req.SortOrder,
		IsVisible: isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(partnerDTO(*row)))
}

func (h *PartnerHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req partnerRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.Partners.Update(c.Request().Context(), id, &model.Partner{
		Name:      req.Name,
		Logo:      normalizeOptString(req.Logo),
		Website:   normalizeOptString(req.Website),
		SortOrder: req.SortOrder,
		IsVisible: isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(partnerDTO(*row)))
}

func (h *PartnerHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	_ = h.Partners.Delete(c.Request().Context(), id)
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func partnerDTO(p model.Partner) map[string]interface{} {
	return map[string]interface{}{
		"id":         p.ID,
		"name":       p.Name,
		"logo":       p.Logo,
		"website":    p.Website,
		"sort_order": p.SortOrder,
		"is_visible": p.IsVisible,
		"created_at": p.CreatedAt,
		"updated_at": p.UpdatedAt,
	}
}


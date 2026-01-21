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

type SystemConfigHandler struct {
	SystemConfigs *service.SystemConfigService
}

func (h *SystemConfigHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	group := c.QueryParam("group")
	rows, err := h.SystemConfigs.List(c.Request().Context(), group)
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		desc := ""
		if row.Description != nil {
			desc = *row.Description
		}
		data = append(data, map[string]interface{}{
			"id":          row.ID,
			"group_name":  row.GroupName,
			"key":         row.Key,
			"value":       row.Value,
			"description": desc,
			"sort_order":  row.SortOrder,
			"is_public":   row.IsPublic,
			"created_at":  row.CreatedAt,
			"updated_at":  row.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type systemConfigRequest struct {
	GroupName   string `json:"group_name" form:"group_name"`
	Key         string `json:"key" form:"key"`
	Value       string `json:"value" form:"value"`
	Description string `json:"description" form:"description"`
	SortOrder   int    `json:"sort_order" form:"sort_order"`
	IsPublic    bool   `json:"is_public" form:"is_public"`
}

func (h *SystemConfigHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req systemConfigRequest
	_ = c.Bind(&req)
	if req.GroupName == "" || req.Key == "" || req.Value == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	desc := req.Description
	row, err := h.SystemConfigs.Create(c.Request().Context(), &model.SystemConfig{
		GroupName:   req.GroupName,
		Key:         req.Key,
		Value:       req.Value,
		Description: &desc,
		SortOrder:   req.SortOrder,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          row.ID,
		"group_name":  row.GroupName,
		"key":         row.Key,
		"value":       row.Value,
		"description": req.Description,
		"sort_order":  row.SortOrder,
		"is_public":   row.IsPublic,
		"created_at":  row.CreatedAt,
		"updated_at":  row.UpdatedAt,
	}))
}

func (h *SystemConfigHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req systemConfigRequest
	_ = c.Bind(&req)
	if req.GroupName == "" || req.Key == "" || req.Value == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	desc := req.Description
	row, err := h.SystemConfigs.Update(c.Request().Context(), id, &model.SystemConfig{
		GroupName:   req.GroupName,
		Key:         req.Key,
		Value:       req.Value,
		Description: &desc,
		SortOrder:   req.SortOrder,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          row.ID,
		"group_name":  row.GroupName,
		"key":         row.Key,
		"value":       row.Value,
		"description": req.Description,
		"sort_order":  row.SortOrder,
		"is_public":   row.IsPublic,
		"created_at":  row.CreatedAt,
		"updated_at":  row.UpdatedAt,
	}))
}

func (h *SystemConfigHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.SystemConfigs.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}


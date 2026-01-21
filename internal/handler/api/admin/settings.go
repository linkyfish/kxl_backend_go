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

type SettingsHandler struct {
	Settings *service.SettingsService
}

func (h *SettingsHandler) ListCategories(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	typ := c.QueryParam("type")
	items, err := h.Settings.ListCategories(c.Request().Context(), typ)
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, row := range items {
		data = append(data, map[string]interface{}{
			"id":         row.ID,
			"name":       row.Name,
			"type":       row.Type,
			"sort_order": row.SortOrder,
			"created_at": row.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type categoryRequest struct {
	Name      string `json:"name" form:"name"`
	Type      string `json:"type" form:"type"`
	SortOrder int    `json:"sort_order" form:"sort_order"`
}

func (h *SettingsHandler) CreateCategory(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req categoryRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Type == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.CreateCategory(c.Request().Context(), req.Name, req.Type, req.SortOrder)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"name":       row.Name,
		"type":       row.Type,
		"sort_order": row.SortOrder,
		"created_at": row.CreatedAt,
	}))
}

func (h *SettingsHandler) UpdateCategory(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req categoryRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Type == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.UpdateCategory(c.Request().Context(), id, req.Name, req.Type, req.SortOrder)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"name":       row.Name,
		"type":       row.Type,
		"sort_order": row.SortOrder,
		"created_at": row.CreatedAt,
	}))
}

func (h *SettingsHandler) DeleteCategory(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Settings.DeleteCategory(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *SettingsHandler) ListTags(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	typ := c.QueryParam("type")
	items, err := h.Settings.ListTags(c.Request().Context(), typ)
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, row := range items {
		data = append(data, map[string]interface{}{
			"id":         row.ID,
			"name":       row.Name,
			"type":       row.Type,
			"created_at": row.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type tagRequest struct {
	Name string `json:"name" form:"name"`
	Type string `json:"type" form:"type"`
}

func (h *SettingsHandler) CreateTag(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req tagRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Type == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.CreateTag(c.Request().Context(), req.Name, req.Type)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"name":       row.Name,
		"type":       row.Type,
		"created_at": row.CreatedAt,
	}))
}

func (h *SettingsHandler) UpdateTag(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req tagRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Type == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.UpdateTag(c.Request().Context(), id, req.Name, req.Type)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"name":       row.Name,
		"type":       row.Type,
		"created_at": row.CreatedAt,
	}))
}

func (h *SettingsHandler) DeleteTag(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Settings.DeleteTag(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type companyInfoRequest struct {
	Name          string `json:"name" form:"name"`
	Description   string `json:"description" form:"description"`
	Phone         string `json:"phone" form:"phone"`
	Email         string `json:"email" form:"email"`
	Address       string `json:"address" form:"address"`
	WorkingHours  string `json:"working_hours" form:"working_hours"`
	MapCoordinates string `json:"map_coordinates" form:"map_coordinates"`
	HeroTitle     string `json:"hero_title" form:"hero_title"`
	HeroSubtitle  string `json:"hero_subtitle" form:"hero_subtitle"`
}

func (h *SettingsHandler) UpdateCompanyInfo(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req companyInfoRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Description == "" || req.Phone == "" || req.Email == "" || req.Address == "" ||
		req.WorkingHours == "" || req.MapCoordinates == "" || req.HeroTitle == "" || req.HeroSubtitle == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	info, err := h.Settings.UpsertCompanyInfo(c.Request().Context(), &model.CompanyInfo{
		Name:           req.Name,
		Description:    req.Description,
		Phone:          req.Phone,
		Email:          req.Email,
		Address:        req.Address,
		WorkingHours:   req.WorkingHours,
		MapCoordinates: req.MapCoordinates,
		HeroTitle:      req.HeroTitle,
		HeroSubtitle:   req.HeroSubtitle,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":                info.ID,
		"name":              info.Name,
		"description":       info.Description,
		"phone":             info.Phone,
		"email":             info.Email,
		"address":           info.Address,
		"working_hours":     info.WorkingHours,
		"map_coordinates":   info.MapCoordinates,
		"hero_title":        info.HeroTitle,
		"hero_subtitle":     info.HeroSubtitle,
		"created_at":        info.CreatedAt,
		"updated_at":        info.UpdatedAt,
		"stats_years":       info.StatsYears,
		"stats_projects":    info.StatsProjects,
		"stats_clients":     info.StatsClients,
		"stats_satisfaction": info.StatsSatisfaction,
	}))
}

func (h *SettingsHandler) ListMilestones(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.Settings.ListMilestones(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, m := range items {
		data = append(data, map[string]interface{}{
			"id":         m.ID,
			"year":       m.Year,
			"content":    m.Content,
			"sort_order": m.SortOrder,
			"created_at": m.CreatedAt,
			"updated_at": m.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type milestoneRequest struct {
	Year      int    `json:"year" form:"year"`
	Content   string `json:"content" form:"content"`
	SortOrder int    `json:"sort_order" form:"sort_order"`
}

func (h *SettingsHandler) CreateMilestone(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req milestoneRequest
	_ = c.Bind(&req)
	if req.Year <= 0 || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.CreateMilestone(c.Request().Context(), req.Year, req.Content, req.SortOrder)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"year":       row.Year,
		"content":    row.Content,
		"sort_order": row.SortOrder,
		"created_at": row.CreatedAt,
		"updated_at": row.UpdatedAt,
	}))
}

func (h *SettingsHandler) UpdateMilestone(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req milestoneRequest
	_ = c.Bind(&req)
	if req.Year <= 0 || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.UpdateMilestone(c.Request().Context(), id, req.Year, req.Content, req.SortOrder)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"year":       row.Year,
		"content":    row.Content,
		"sort_order": row.SortOrder,
		"created_at": row.CreatedAt,
		"updated_at": row.UpdatedAt,
	}))
}

func (h *SettingsHandler) DeleteMilestone(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Settings.DeleteMilestone(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *SettingsHandler) ListTeam(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.Settings.ListTeamMembers(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, m := range items {
		data = append(data, map[string]interface{}{
			"id":         m.ID,
			"name":       m.Name,
			"title":      m.Title,
			"avatar":     m.Avatar,
			"bio":        m.Bio,
			"sort_order": m.SortOrder,
			"created_at": m.CreatedAt,
			"updated_at": m.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type teamMemberRequest struct {
	Name      string `json:"name" form:"name"`
	Title     string `json:"title" form:"title"`
	Avatar    string `json:"avatar" form:"avatar"`
	Bio       string `json:"bio" form:"bio"`
	SortOrder int    `json:"sort_order" form:"sort_order"`
}

func (h *SettingsHandler) CreateTeamMember(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req teamMemberRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Title == "" || req.Avatar == "" || req.Bio == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.CreateTeamMember(c.Request().Context(), &model.TeamMember{
		Name:      req.Name,
		Title:     req.Title,
		Avatar:    req.Avatar,
		Bio:       req.Bio,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"name":       row.Name,
		"title":      row.Title,
		"avatar":     row.Avatar,
		"bio":        row.Bio,
		"sort_order": row.SortOrder,
		"created_at": row.CreatedAt,
		"updated_at": row.UpdatedAt,
	}))
}

func (h *SettingsHandler) UpdateTeamMember(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req teamMemberRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Title == "" || req.Avatar == "" || req.Bio == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row, err := h.Settings.UpdateTeamMember(c.Request().Context(), id, &model.TeamMember{
		Name:      req.Name,
		Title:     req.Title,
		Avatar:    req.Avatar,
		Bio:       req.Bio,
		SortOrder: req.SortOrder,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"name":       row.Name,
		"title":      row.Title,
		"avatar":     row.Avatar,
		"bio":        row.Bio,
		"sort_order": row.SortOrder,
		"created_at": row.CreatedAt,
		"updated_at": row.UpdatedAt,
	}))
}

func (h *SettingsHandler) DeleteTeamMember(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Settings.DeleteTeamMember(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *SettingsHandler) DashboardStats(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "dashboard:view"); err != nil {
		return err
	}
	stats, err := h.Settings.DashboardStats(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(stats))
}


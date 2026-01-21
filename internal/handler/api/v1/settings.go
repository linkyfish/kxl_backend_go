package v1

import (
	"net/http"

	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type SettingsHandler struct {
	Settings *service.SettingsService
}

func (h *SettingsHandler) GetCompanyInfo(c echo.Context) error {
	info, err := h.Settings.GetCompanyInfo(c.Request().Context())
	if err != nil {
		return err
	}
	if info == nil {
		created, err := h.Settings.UpsertCompanyInfo(c.Request().Context(), &model.CompanyInfo{
			Name:          "Company",
			Description:   "",
			Phone:         "",
			Email:         "",
			Address:       "",
			WorkingHours:  "",
			MapCoordinates: "",
			HeroTitle:     "",
			HeroSubtitle:  "",
			StatsYears:    nil,
			StatsProjects: nil,
			StatsClients:  nil,
			StatsSatisfaction: nil,
		})
		if err != nil {
			return err
		}
		info = created
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

func (h *SettingsHandler) ListTeam(c echo.Context) error {
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


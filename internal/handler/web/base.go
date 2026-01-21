package web

import (
	"context"
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

// BaseData contains common data shared across SSR pages.
type BaseData struct {
	Company       *model.CompanyInfo
	FriendlyLinks []model.FriendlyLink
}

func LoadBaseData(ctx context.Context, settings *service.SettingsService, friendly *service.FriendlyLinkService) (BaseData, error) {
	var out BaseData
	if settings != nil {
		company, err := settings.GetCompanyInfo(ctx)
		if err != nil {
			return out, err
		}
		out.Company = company
	}
	if friendly != nil {
		links, err := friendly.ListVisible(ctx)
		if err != nil {
			return out, err
		}
		out.FriendlyLinks = links
	}
	return out, nil
}

// InjectBaseContext populates common template vars.
func InjectBaseContext(dst pongo2.Context, c echo.Context, base BaseData) {
	if dst == nil {
		return
	}

	dst["company"] = companyDTO(base.Company)
	dst["friendly_links"] = friendlyLinkDTOs(base.FriendlyLinks)

	// Convenience URLs for meta tags.
	if c != nil && c.Request() != nil {
		req := c.Request()
		scheme := "http"
		if req.TLS != nil {
			scheme = "https"
		}
		if xfProto := strings.TrimSpace(req.Header.Get("X-Forwarded-Proto")); xfProto != "" {
			scheme = xfProto
		}
		host := strings.TrimSpace(req.Host)
		if host != "" {
			baseURL := scheme + "://" + host
			dst["base_url"] = baseURL
			if req.URL != nil {
				dst["current_url"] = baseURL + req.URL.RequestURI()
			}
		}
	}
}

func companyDTO(info *model.CompanyInfo) map[string]interface{} {
	// Use map keys matching template usage (snake_case) to avoid Pongo2 reflection surprises.
	out := map[string]interface{}{
		"id":               0,
		"name":             "",
		"description":      "",
		"phone":            "",
		"email":            "",
		"address":          "",
		"working_hours":    "",
		"map_coordinates":  "",
		"hero_title":       "",
		"hero_subtitle":    "",
		"stats_years":      nil,
		"stats_projects":   nil,
		"stats_clients":    nil,
		"stats_satisfaction": nil,

		// Optional fields referenced by templates but not present in DB schema/model.
		"logo":          nil,
		"slogan":        nil,
		"tagline":       nil,
		"hero_highlight": nil,
		"keywords":      nil,
		"favicon":       nil,
		"wechat":        nil,
		"weibo":         nil,
		"icp":           nil,
		"police_record": nil,
		"about":         nil,
		"about_image":   nil,
	}

	if info == nil {
		return out
	}

	out["id"] = info.ID
	out["name"] = info.Name
	out["description"] = info.Description
	out["phone"] = info.Phone
	out["email"] = info.Email
	out["address"] = info.Address
	out["working_hours"] = info.WorkingHours
	out["map_coordinates"] = info.MapCoordinates
	out["hero_title"] = info.HeroTitle
	out["hero_subtitle"] = info.HeroSubtitle
	out["stats_years"] = info.StatsYears
	out["stats_projects"] = info.StatsProjects
	out["stats_clients"] = info.StatsClients
	out["stats_satisfaction"] = info.StatsSatisfaction
	return out
}

func friendlyLinkDTOs(rows []model.FriendlyLink) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		out = append(out, map[string]interface{}{
			"id":          r.ID,
			"name":        r.Name,
			"url":         r.URL,
			"logo":        r.Logo,
			"description": r.Description,
			"sort_order":  r.SortOrder,
		})
	}
	return out
}

func wantsJSON(req *http.Request) bool {
	if req == nil {
		return false
	}
	accept := req.Header.Get("Accept")
	return strings.Contains(accept, "application/json")
}


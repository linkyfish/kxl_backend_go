package web

import (
	"net/http"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type AboutHandler struct {
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
}

func (h *AboutHandler) Index(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	ctx := pongo2.Context{
		"page_title": "关于我们",
		"breadcrumbs": []map[string]interface{}{
			{"title": "关于我们", "url": "/about"},
		},
	}
	InjectBaseContext(ctx, c, base)

	if h.Settings != nil {
		milestones, err := h.Settings.ListMilestones(c.Request().Context())
		if err != nil {
			return err
		}
		ctx["milestones"] = milestoneTimelineItems(milestones)

		team, err := h.Settings.ListTeamMembers(c.Request().Context())
		if err != nil {
			return err
		}
		ctx["team_members"] = teamMemberDTOs(team)
	}

	return c.Render(http.StatusOK, "pages/about.html", ctx)
}

func milestoneTimelineItems(rows []model.Milestone) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, m := range rows {
		// timeline.html expects {year/date, title, description}; milestone table only has content.
		out = append(out, map[string]interface{}{
			"id":          m.ID,
			"year":        m.Year,
			"title":       m.Content,
			"description": "",
		})
	}
	return out
}

func teamMemberDTOs(rows []model.TeamMember) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, m := range rows {
		out = append(out, map[string]interface{}{
			"id":       m.ID,
			"name":     m.Name,
			"avatar":   m.Avatar,
			"bio":      m.Bio,
			"role":     m.Title,
			"position": m.Title,
		})
	}
	return out
}


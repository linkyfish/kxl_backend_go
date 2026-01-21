package web

import (
	"net/http"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type HomeHandler struct {
	Settings     *service.SettingsService
	Banners      *service.BannerService
	Projects     *service.ProjectService
	Cases        *service.CaseService
	Articles     *service.ArticleService
	Testimonials *service.TestimonialService
	Solutions    *service.SolutionService
	Partners     *service.PartnerService
	Friendly     *service.FriendlyLinkService
}

func (h *HomeHandler) Index(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	ctx := pongo2.Context{
		"page_title": "首页",
	}
	InjectBaseContext(ctx, c, base)

	// Stats (normalize values to avoid double "+/%" when templates append them).
	if base.Company != nil {
		ctx["stats"] = map[string]interface{}{
			"years":        normalizeStat(base.Company.StatsYears, "8"),
			"projects":     normalizeStat(base.Company.StatsProjects, "100"),
			"clients":      normalizeStat(base.Company.StatsClients, "200"),
			"satisfaction": normalizeStat(base.Company.StatsSatisfaction, "98"),
		}
	}

	// Banners.
	if h.Banners != nil {
		banners, err := h.Banners.ListVisible(c.Request().Context())
		if err != nil {
			return err
		}
		ctx["banners"] = bannerDTOs(banners)
	}

	// Featured projects.
	if h.Projects != nil {
		rows, _, err := h.Projects.ListPublic(c.Request().Context(), 1, 6, nil, "")
		if err != nil {
			return err
		}
		items, err := buildProjectListItems(c.Request().Context(), rows, h.Projects)
		if err != nil {
			return err
		}
		ctx["featured_projects"] = items
	}

	// Featured cases.
	if h.Cases != nil {
		rows, _, err := h.Cases.ListPublic(c.Request().Context(), 1, 3, nil, "")
		if err != nil {
			return err
		}
		items, err := buildCaseListItems(c.Request().Context(), rows, h.Cases)
		if err != nil {
			return err
		}
		ctx["featured_cases"] = items
	}

	// Latest articles.
	if h.Articles != nil {
		rows, _, err := h.Articles.ListPublic(c.Request().Context(), 1, 3, nil, "")
		if err != nil {
			return err
		}
		items, err := buildArticleListItems(c.Request().Context(), rows, h.Articles)
		if err != nil {
			return err
		}
		ctx["latest_articles"] = items
	}

	// Testimonials / solutions / partners.
	if h.Testimonials != nil {
		rows, err := h.Testimonials.ListVisible(c.Request().Context())
		if err != nil {
			return err
		}
		ctx["testimonials"] = testimonialDTOs(rows)
	}
	if h.Solutions != nil {
		rows, err := h.Solutions.ListVisible(c.Request().Context())
		if err != nil {
			return err
		}
		ctx["solutions"] = solutionDTOs(rows)
	}
	if h.Partners != nil {
		rows, err := h.Partners.ListVisible(c.Request().Context())
		if err != nil {
			return err
		}
		ctx["partners"] = partnerDTOs(rows)
	}

	return c.Render(http.StatusOK, "pages/home.html", ctx)
}

func normalizeStat(raw *string, fallback string) string {
	if raw == nil {
		return fallback
	}
	s := strings.TrimSpace(*raw)
	s = strings.TrimSuffix(s, "+")
	s = strings.TrimSuffix(s, "%")
	if s == "" {
		return fallback
	}
	return s
}

func bannerDTOs(rows []model.Banner) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, b := range rows {
		out = append(out, map[string]interface{}{
			"id":         b.ID,
			"title":      b.Title,
			"subtitle":   b.Subtitle,
			"highlight":  b.Highlight,
			"tag":        b.Tag,
			"image":      b.Image,
			"link":       b.Link,
			"link_text":  b.LinkText,
			"bg_class":   b.BgClass,
			"sort_order": b.SortOrder,
		})
	}
	return out
}

func testimonialDTOs(rows []model.Testimonial) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, t := range rows {
		out = append(out, map[string]interface{}{
			"id":         t.ID,
			"name":       t.Name,
			"title":      t.Title,
			"company":    t.Company,
			"avatar":     t.Avatar,
			"content":    t.Content,
			"rating":     t.Rating,
			"sort_order": t.SortOrder,
		})
	}
	return out
}

func solutionDTOs(rows []model.Solution) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, s := range rows {
		out = append(out, map[string]interface{}{
			"id":          s.ID,
			"name":        s.Name,
			"description": s.Description,
			"icon":        s.Icon,
			"bg_class":    s.BgClass,
			"link":        s.Link,
			"sort_order":  s.SortOrder,
		})
	}
	return out
}

func partnerDTOs(rows []model.Partner) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, p := range rows {
		out = append(out, map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"logo":       p.Logo,
			"website":    p.Website,
			"sort_order": p.SortOrder,
		})
	}
	return out
}


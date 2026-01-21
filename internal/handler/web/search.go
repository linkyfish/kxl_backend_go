package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type SearchHandler struct {
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
	Search   *service.SearchService
}

func (h *SearchHandler) Index(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	keyword := strings.TrimSpace(c.QueryParam("q"))
	searchType := strings.TrimSpace(c.QueryParam("type"))
	page := int64(1)
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			page = n
		}
	}
	pageSize := int64(20)

	results := []map[string]interface{}{}
	var total int64 = 0

	if keyword != "" && h.Search != nil {
		var typPtr *string
		if searchType != "" {
			typPtr = &searchType
		}

		items, t, err := h.Search.Search(c.Request().Context(), keyword, typPtr, page, pageSize)
		if err != nil {
			return err
		}
		total = t

		for _, it := range items {
			results = append(results, map[string]interface{}{
				"id":      it.ID,
				"type":    it.Type,
				"title":   it.Title,
				"url":     it.URL,
				"excerpt": it.Summary,
				"date":    "",
			})
		}
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	query := ""
	if keyword != "" {
		query = "q=" + keyword
		if searchType != "" {
			query += "&type=" + searchType
		}
	}

	ctx := pongo2.Context{
		"page_title":  "搜索结果",
		"breadcrumbs": []map[string]interface{}{{"title": "搜索结果", "url": "/search"}},
		"keyword":     keyword,
		"search_type": searchType,
		"results":     results,
		"total":       total,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  total,
			"base_url":     "/search",
			"query":        query,
		},
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/search.html", ctx)
}


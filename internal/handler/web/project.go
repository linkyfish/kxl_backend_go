package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	DB       *gorm.DB
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
	Projects *service.ProjectService
}

func (h *ProjectHandler) List(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	page := int64(1)
	if raw := strings.TrimSpace(c.QueryParam("page")); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			page = n
		}
	}
	pageSize := int64(12)

	var categoryID *int
	if raw := strings.TrimSpace(c.QueryParam("category")); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			categoryID = &n
		}
	}
	keyword := strings.TrimSpace(c.QueryParam("q"))

	rows, total, err := h.Projects.ListPublic(c.Request().Context(), page, pageSize, categoryID, keyword)
	if err != nil {
		return err
	}
	projects, err := buildProjectListItems(c.Request().Context(), rows, h.Projects)
	if err != nil {
		return err
	}

	// Categories
	categories := []map[string]interface{}{}
	if h.Settings != nil {
		cats, err := h.Settings.ListCategories(c.Request().Context(), "project")
		if err == nil {
			categories = categoryDTOs(cats)
		}
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	var currentCategory interface{} = nil
	if categoryID != nil {
		currentCategory = *categoryID
	}

	ctx := pongo2.Context{
		"page_title":       "软件产品",
		"breadcrumbs":      []map[string]interface{}{{"title": "软件产品", "url": "/projects"}},
		"projects":         projects,
		"categories":       categories,
		"current_category": currentCategory,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  total,
			"base_url":     "/projects",
			"query":        "",
		},
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/projects/list.html", ctx)
}

func (h *ProjectHandler) Detail(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	p, err := h.Projects.GetPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}

	features, err := h.Projects.ListFeatures(c.Request().Context(), p.ID)
	if err != nil {
		return err
	}
	featureDTOs := make([]map[string]interface{}, 0, len(features))
	for _, f := range features {
		featureDTOs = append(featureDTOs, map[string]interface{}{
			"id":          f.ID,
			"name":        f.Name,
			"description": f.Description,
			"icon":        f.Icon,
			"sort_order":  f.SortOrder,
		})
	}

	media, err := h.Projects.ListMedia(c.Request().Context(), p.ID)
	if err != nil {
		return err
	}
	mediaDTOs := make([]map[string]interface{}, 0, len(media))
	for _, m := range media {
		mediaDTOs = append(mediaDTOs, map[string]interface{}{
			"id":         m.ID,
			"type":       m.Type,
			"url":        m.URL,
			"title":      m.Title,
			"sort_order": m.SortOrder,
		})
	}

	versions, err := h.Projects.ListVersions(c.Request().Context(), p.ID)
	if err != nil {
		return err
	}
	versionDTOs := make([]map[string]interface{}, 0, len(versions))
	for _, v := range versions {
		date := ""
		if !v.ReleaseDate.IsZero() {
			date = v.ReleaseDate.UTC().Format("2006-01-02")
		}
		versionDTOs = append(versionDTOs, map[string]interface{}{
			"id":           v.ID,
			"version":      v.Version,
			"release_date": date,
			"changelog":    v.Changelog,
		})
	}

	// Category
	var category interface{} = nil
	if p.CategoryID != nil && h.DB != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *p.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
	}

	// Tags
	tagsByProject, err := h.Projects.LoadTagsByProjectIDs(c.Request().Context(), []string{p.ID})
	if err != nil {
		return err
	}
	tags := []map[string]interface{}{}
	for _, t := range tagsByProject[p.ID] {
		tags = append(tags, tagDTO(t))
	}

	project := map[string]interface{}{
		"id":          p.ID,
		"name":        p.Name,
		"description": p.Description,
		"cover_image": p.CoverImage,
		"status":      p.Status,
		"sort_order":  p.SortOrder,
		"category":    category,
		"tags":        tags,
		"features":    featureDTOs,
		"media":       mediaDTOs,
		"versions":    versionDTOs,
		"created_at":  p.CreatedAt,
		"updated_at":  p.UpdatedAt,

		// Optional template fields not present in schema.
		"brief":      nil,
		"demo_url":   nil,
		"video_url":  nil,
		"platform":   nil,
		"tech_stack": nil,
	}

	ctx := pongo2.Context{
		"page_title":  p.Name,
		"breadcrumbs": []map[string]interface{}{{"title": "软件产品", "url": "/projects"}, {"title": p.Name, "url": ""}},
		"project":     project,
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/projects/detail.html", ctx)
}


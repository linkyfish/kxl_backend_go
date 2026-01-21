package web

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CaseHandler struct {
	DB       *gorm.DB
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
	Cases    *service.CaseService
	Projects *service.ProjectService
}

func (h *CaseHandler) List(c echo.Context) error {
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

	rows, total, err := h.Cases.ListPublic(c.Request().Context(), page, pageSize, categoryID, keyword)
	if err != nil {
		return err
	}
	cases, err := buildCaseListItems(c.Request().Context(), rows, h.Cases)
	if err != nil {
		return err
	}

	// Categories
	categories := []map[string]interface{}{}
	if h.Settings != nil {
		cats, err := h.Settings.ListCategories(c.Request().Context(), "case")
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
		"page_title":       "成功案例",
		"breadcrumbs":      []map[string]interface{}{{"title": "成功案例", "url": "/cases"}},
		"cases":            cases,
		"categories":       categories,
		"current_category": currentCategory,
		"pagination": map[string]interface{}{
			"current_page": page,
			"total_pages":  totalPages,
			"total_items":  total,
			"base_url":     "/cases",
			"query":        "",
		},
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/cases/list.html", ctx)
}

func (h *CaseHandler) Detail(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	cs, err := h.Cases.GetPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// Category
	var category interface{} = nil
	if cs.CategoryID != nil && h.DB != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *cs.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
	}

	// Related projects
	relatedProjects := []map[string]interface{}{}
	if h.DB != nil && h.Projects != nil {
		projectIDs, err := h.Cases.LoadProjectIDsByCaseID(c.Request().Context(), cs.ID)
		if err != nil {
			return err
		}
		if len(projectIDs) > 0 {
			var projects []model.Project
			if err := h.DB.WithContext(c.Request().Context()).
				Where("id in ?", projectIDs).
				Where("status = ?", 1).
				Order("sort_order asc").Order("id asc").
				Find(&projects).Error; err != nil {
				return kxlerrors.Internal("db error")
			}

			// Load category/tag maps for related projects.
			catIDs := make([]int, 0)
			seen := make(map[int]struct{})
			pids := make([]string, 0, len(projects))
			for _, p := range projects {
				pids = append(pids, p.ID)
				if p.CategoryID != nil {
					if _, ok := seen[*p.CategoryID]; !ok {
						seen[*p.CategoryID] = struct{}{}
						catIDs = append(catIDs, *p.CategoryID)
					}
				}
			}
			projectCategoryMap, err := h.Projects.LoadCategoryMap(c.Request().Context(), catIDs)
			if err != nil {
				return err
			}
			tagsByProject, err := h.Projects.LoadTagsByProjectIDs(c.Request().Context(), pids)
			if err != nil {
				return err
			}

			for _, p := range projects {
				var pCategory interface{} = nil
				if p.CategoryID != nil {
					if cat, ok := projectCategoryMap[*p.CategoryID]; ok {
						pCategory = categoryDTO(cat)
					}
				}
				tags := []map[string]interface{}{}
				for _, t := range tagsByProject[p.ID] {
					tags = append(tags, tagDTO(t))
				}
				relatedProjects = append(relatedProjects, map[string]interface{}{
					"id":          p.ID,
					"name":        p.Name,
					"description": p.Description,
					"cover_image": p.CoverImage,
					"status":      p.Status,
					"sort_order":  p.SortOrder,
					"category":    pCategory,
					"tags":        tags,
				})
			}
		}
	}

	results := decodeJSONArray(cs.Results)

	caseObj := map[string]interface{}{
		"id":                 cs.ID,
		"client_name":        cs.ClientName,
		"client_logo":        nil,
		"cover_image":        cs.CoverImage,
		"summary":            cs.Summary,
		"background":         cs.Background,
		"solution":           cs.Solution,
		"results":            results,
		"testimonial":        cs.Testimonial,
		"testimonial_author": cs.TestimonialAuthor,
		"testimonial_title":  cs.TestimonialTitle,
		"status":             cs.Status,
		"category":           category,
		"related_projects":   relatedProjects,
		"created_at":         cs.CreatedAt,
		"updated_at":         cs.UpdatedAt,
	}

	ctx := pongo2.Context{
		"page_title":  cs.ClientName,
		"breadcrumbs": []map[string]interface{}{{"title": "成功案例", "url": "/cases"}, {"title": cs.ClientName, "url": ""}},
		"case":        caseObj,
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/cases/detail.html", ctx)
}

func decodeJSONArray(raw datatypes.JSON) []interface{} {
	if len(raw) == 0 {
		return []interface{}{}
	}
	var out []interface{}
	if json.Unmarshal(raw, &out) == nil {
		return out
	}
	// Some deployments store results as a single JSON string.
	var single string
	if json.Unmarshal(raw, &single) == nil && single != "" {
		return []interface{}{single}
	}
	return []interface{}{}
}


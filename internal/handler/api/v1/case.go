package v1

import (
	"net/http"
	"strconv"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type CaseHandler struct {
	DB       *gorm.DB
	Cases    *service.CaseService
	Projects *service.ProjectService
}

func (h *CaseHandler) List(c echo.Context) error {
	page := int64(1)
	pageSize := int64(10)
	if raw := c.QueryParam("page"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			page = n
		}
	}
	if raw := c.QueryParam("page_size"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			pageSize = n
		}
	}
	if page < 1 {
		return kxlerrors.Validation("validation error: page must be >= 1")
	}
	if pageSize < 1 || pageSize > 200 {
		return kxlerrors.Validation("validation error: page_size must be between 1 and 200")
	}

	var categoryID *int
	if raw := c.QueryParam("category_id"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			categoryID = &n
		}
	}
	keyword := c.QueryParam("keyword")

	rows, total, err := h.Cases.ListPublic(c.Request().Context(), page, pageSize, categoryID, keyword)
	if err != nil {
		return err
	}

	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	for _, row := range rows {
		if row.CategoryID != nil {
			if _, ok := seenCat[*row.CategoryID]; !ok {
				seenCat[*row.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *row.CategoryID)
			}
		}
	}
	categoryMap, err := h.Cases.LoadCategoryMap(c.Request().Context(), categoryIDs)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		var category interface{} = nil
		if row.CategoryID != nil {
			if cat, ok := categoryMap[*row.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}
		items = append(items, map[string]interface{}{
			"id":          row.ID,
			"client_name": row.ClientName,
			"cover_image": row.CoverImage,
			"summary":     row.Summary,
			"status":      row.Status,
			"category":    category,
		})
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"items":       items,
		"total":       total,
		"page":        page,
		"page_size":   pageSize,
		"total_pages": totalPages,
	}))
}

func (h *CaseHandler) Detail(c echo.Context) error {
	id := c.Param("id")
	cs, err := h.Cases.GetPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}

	var category interface{} = nil
	if cs.CategoryID != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *cs.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
	}

	projectIDs, err := h.Cases.LoadProjectIDsByCaseID(c.Request().Context(), cs.ID)
	if err != nil {
		return err
	}

	relatedProjects := []map[string]interface{}{}
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

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":                 cs.ID,
		"client_name":        cs.ClientName,
		"cover_image":        cs.CoverImage,
		"summary":            cs.Summary,
		"background":         cs.Background,
		"solution":           cs.Solution,
		"results":            cs.Results,
		"testimonial":        cs.Testimonial,
		"testimonial_author": cs.TestimonialAuthor,
		"testimonial_title":  cs.TestimonialTitle,
		"status":             cs.Status,
		"category":           category,
		"related_projects":   relatedProjects,
		"created_at":         cs.CreatedAt,
		"updated_at":         cs.UpdatedAt,
	}))
}


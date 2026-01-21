package v1

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"gorm.io/gorm"
)

type ProjectHandler struct {
	DB       *gorm.DB
	Projects *service.ProjectService
}

func (h *ProjectHandler) List(c echo.Context) error {
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

	rows, total, err := h.Projects.ListPublic(c.Request().Context(), page, pageSize, categoryID, keyword)
	if err != nil {
		return err
	}

	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	projectIDs := make([]string, 0, len(rows))
	for _, p := range rows {
		projectIDs = append(projectIDs, p.ID)
		if p.CategoryID != nil {
			if _, ok := seenCat[*p.CategoryID]; !ok {
				seenCat[*p.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *p.CategoryID)
			}
		}
	}
	categoryMap, err := h.Projects.LoadCategoryMap(c.Request().Context(), categoryIDs)
	if err != nil {
		return err
	}
	tagsByProject, err := h.Projects.LoadTagsByProjectIDs(c.Request().Context(), projectIDs)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, p := range rows {
		var category interface{} = nil
		if p.CategoryID != nil {
			if cat, ok := categoryMap[*p.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}
		tags := []map[string]interface{}{}
		for _, t := range tagsByProject[p.ID] {
			tags = append(tags, tagDTO(t))
		}
		items = append(items, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"cover_image": p.CoverImage,
			"status":      p.Status,
			"sort_order":  p.SortOrder,
			"category":    category,
			"tags":        tags,
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

func (h *ProjectHandler) Detail(c echo.Context) error {
	id := c.Param("id")
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

	var category interface{} = nil
	if p.CategoryID != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *p.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
	}

	tagsByProject, err := h.Projects.LoadTagsByProjectIDs(c.Request().Context(), []string{p.ID})
	if err != nil {
		return err
	}
	tags := []map[string]interface{}{}
	for _, t := range tagsByProject[p.ID] {
		tags = append(tags, tagDTO(t))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
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
	}))
}

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

type ArticleHandler struct {
	DB       *gorm.DB
	Articles *service.ArticleService
}

func (h *ArticleHandler) List(c echo.Context) error {
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

	rows, total, err := h.Articles.ListPublic(c.Request().Context(), page, pageSize, categoryID, keyword)
	if err != nil {
		return err
	}

	// Category map
	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	articleIDs := make([]string, 0, len(rows))
	for _, a := range rows {
		articleIDs = append(articleIDs, a.ID)
		if a.CategoryID != nil {
			if _, ok := seenCat[*a.CategoryID]; !ok {
				seenCat[*a.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *a.CategoryID)
			}
		}
	}
	categoryMap, err := h.Articles.LoadCategoryMap(c.Request().Context(), categoryIDs)
	if err != nil {
		return err
	}
	tagsByArticle, err := h.Articles.LoadTagsByArticleIDs(c.Request().Context(), articleIDs)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, a := range rows {
		var category interface{} = nil
		if a.CategoryID != nil {
			if cat, ok := categoryMap[*a.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}

		tags := []map[string]interface{}{}
		for _, t := range tagsByArticle[a.ID] {
			tags = append(tags, tagDTO(t))
		}

		items = append(items, map[string]interface{}{
			"id":           a.ID,
			"title":        a.Title,
			"summary":      a.Summary,
			"cover_image":  a.CoverImage,
			"category":     category,
			"published_at": a.PublishedAt,
			"view_count":   a.ViewCount,
			"tags":         tags,
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

func (h *ArticleHandler) Detail(c echo.Context) error {
	id := c.Param("id")
	a, err := h.Articles.GetPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}
	oldViewCount := a.ViewCount
	_ = h.Articles.IncrementViewCount(c.Request().Context(), id)

	var category interface{} = nil
	if a.CategoryID != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *a.CategoryID).First(&cat).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return kxlerrors.NotFound("not found: category not found")
			}
			return kxlerrors.Internal("db error")
		}
		category = categoryDTO(cat)
	}

	tagsByArticle, err := h.Articles.LoadTagsByArticleIDs(c.Request().Context(), []string{a.ID})
	if err != nil {
		return err
	}
	tags := []map[string]interface{}{}
	for _, t := range tagsByArticle[a.ID] {
		tags = append(tags, tagDTO(t))
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":           a.ID,
		"title":        a.Title,
		"summary":      a.Summary,
		"content":      a.Content,
		"cover_image":  a.CoverImage,
		"category":     category,
		"published_at": a.PublishedAt,
		"view_count":   oldViewCount + 1,
		"tags":         tags,
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
	}))
}

func (h *ArticleHandler) Related(c echo.Context) error {
	id := c.Param("id")
	rows, err := h.Articles.RelatedPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}
	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	articleIDs := make([]string, 0, len(rows))
	for _, a := range rows {
		articleIDs = append(articleIDs, a.ID)
		if a.CategoryID != nil {
			if _, ok := seenCat[*a.CategoryID]; !ok {
				seenCat[*a.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *a.CategoryID)
			}
		}
	}
	categoryMap, err := h.Articles.LoadCategoryMap(c.Request().Context(), categoryIDs)
	if err != nil {
		return err
	}
	tagsByArticle, err := h.Articles.LoadTagsByArticleIDs(c.Request().Context(), articleIDs)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, a := range rows {
		var category interface{} = nil
		if a.CategoryID != nil {
			if cat, ok := categoryMap[*a.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}
		tags := []map[string]interface{}{}
		for _, t := range tagsByArticle[a.ID] {
			tags = append(tags, tagDTO(t))
		}
		items = append(items, map[string]interface{}{
			"id":           a.ID,
			"title":        a.Title,
			"summary":      a.Summary,
			"cover_image":  a.CoverImage,
			"category":     category,
			"published_at": a.PublishedAt,
			"view_count":   a.ViewCount,
			"tags":         tags,
		})
	}
	return c.JSON(http.StatusOK, response.Success(items))
}

func (h *ArticleHandler) Navigation(c echo.Context) error {
	id := c.Param("id")
	prev, next, err := h.Articles.NavigationPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"previous": prev,
		"next":     next,
	}))
}

func categoryDTO(c model.Category) map[string]interface{} {
	return map[string]interface{}{
		"id":         c.ID,
		"name":       c.Name,
		"type":       c.Type,
		"sort_order": c.SortOrder,
		"created_at": c.CreatedAt,
	}
}

func tagDTO(t model.Tag) map[string]interface{} {
	return map[string]interface{}{
		"id":         t.ID,
		"name":       t.Name,
		"type":       t.Type,
		"created_at": t.CreatedAt,
	}
}


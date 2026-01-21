package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"gorm.io/gorm"
)

type ArticleHandler struct {
	DB       *gorm.DB
	Articles *service.ArticleService
}

func (h *ArticleHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "articles:read"); err != nil {
		return err
	}

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

	q := h.DB.WithContext(c.Request().Context()).Model(&model.Article{})
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(title ILIKE ? OR summary ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	var rows []model.Article
	if err := q.Order("published_at desc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	// Category & tags.
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
	if err := middleware.AdminRequirePermission(c, "articles:read"); err != nil {
		return err
	}
	id := c.Param("id")

	var a model.Article
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: article not found")
		}
		return kxlerrors.Internal("db error")
	}

	var category interface{} = nil
	if a.CategoryID != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *a.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
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
		"view_count":   a.ViewCount,
		"tags":         tags,
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
	}))
}

type articleUpsertRequest struct {
	Title      string  `json:"title" form:"title"`
	Summary    string  `json:"summary" form:"summary"`
	Content    string  `json:"content" form:"content"`
	CoverImage *string `json:"cover_image" form:"cover_image"`
	CategoryID *int    `json:"category_id" form:"category_id"`
}

func (h *ArticleHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "articles:write"); err != nil {
		return err
	}
	var req articleUpsertRequest
	_ = c.Bind(&req)
	if req.Title == "" || req.Summary == "" || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	a := &model.Article{
		Title:      req.Title,
		Summary:    req.Summary,
		Content:    req.Content,
		CoverImage: normalizeOptString(req.CoverImage),
		CategoryID: req.CategoryID,
		ViewCount:  0,
		Status:     0,
		PublishedAt: nil,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(a).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":           a.ID,
		"title":        a.Title,
		"summary":      a.Summary,
		"content":      a.Content,
		"cover_image":  a.CoverImage,
		"category":     nil,
		"published_at": a.PublishedAt,
		"view_count":   a.ViewCount,
		"tags":         []interface{}{},
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
	}))
}

func (h *ArticleHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "articles:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req articleUpsertRequest
	_ = c.Bind(&req)
	if req.Title == "" || req.Summary == "" || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var a model.Article
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: article not found")
		}
		return kxlerrors.Internal("db error")
	}

	a.Title = req.Title
	a.Summary = req.Summary
	a.Content = req.Content
	a.CoverImage = normalizeOptString(req.CoverImage)
	a.CategoryID = req.CategoryID

	if err := h.DB.WithContext(c.Request().Context()).Save(&a).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return h.Detail(c)
}

func (h *ArticleHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "articles:write"); err != nil {
		return err
	}
	id := c.Param("id")
	res := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).Delete(&model.Article{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: article not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type articleStatusRequest struct {
	Status *int16 `json:"status" form:"status"`
}

func (h *ArticleHandler) UpdateStatus(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "articles:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req articleStatusRequest
	_ = c.Bind(&req)
	if req.Status == nil {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var a model.Article
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&a).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: article not found")
		}
		return kxlerrors.Internal("db error")
	}

	a.Status = *req.Status
	if a.Status == 1 {
		now := time.Now().UTC()
		a.PublishedAt = &now
	} else {
		a.PublishedAt = nil
	}
	if err := h.DB.WithContext(c.Request().Context()).Save(&a).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return h.Detail(c)
}

type articleTagsRequest struct {
	TagIDs []int `json:"tag_ids" form:"tag_ids"`
}

func (h *ArticleHandler) SetTags(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "articles:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req articleTagsRequest
	_ = c.Bind(&req)

	// Best-effort: allow empty list to clear tags.
	unique := make([]int, 0, len(req.TagIDs))
	seen := make(map[int]struct{})
	for _, tid := range req.TagIDs {
		if tid <= 0 {
			continue
		}
		if _, ok := seen[tid]; ok {
			continue
		}
		seen[tid] = struct{}{}
		unique = append(unique, tid)
	}

	if err := h.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("article_id = ?", id).Delete(&model.ArticleTag{}).Error; err != nil {
			return kxlerrors.Internal("db error")
		}
		for _, tid := range unique {
			if err := tx.Create(&model.ArticleTag{ArticleID: id, TagID: tid}).Error; err != nil {
				return kxlerrors.Internal("db error")
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

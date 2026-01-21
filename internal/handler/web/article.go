package web

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/flosch/pongo2/v6"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type ArticleHandler struct {
	DB       *gorm.DB
	Settings *service.SettingsService
	Friendly *service.FriendlyLinkService
	Articles *service.ArticleService
}

func (h *ArticleHandler) List(c echo.Context) error {
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
	view := strings.TrimSpace(c.QueryParam("view"))
	if view == "" {
		view = "grid"
	}

	rows, total, err := h.Articles.ListPublic(c.Request().Context(), page, pageSize, categoryID, keyword)
	if err != nil {
		return err
	}
	articles, err := buildArticleListItems(c.Request().Context(), rows, h.Articles)
	if err != nil {
		return err
	}

	// Categories
	categories := []map[string]interface{}{}
	if h.Settings != nil {
		cats, err := h.Settings.ListCategories(c.Request().Context(), "article")
		if err == nil {
			categories = categoryDTOs(cats)
		}
	}

	totalPages := int64(0)
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	pagination := map[string]interface{}{
		"current_page": page,
		"total_pages":  totalPages,
		"total_items":  total,
		"base_url":     "/articles",
		"query":        "",
	}

	var currentCategory interface{} = nil
	if categoryID != nil {
		currentCategory = *categoryID
	}

	ctx := pongo2.Context{
		"page_title":       "新闻动态",
		"breadcrumbs":      []map[string]interface{}{{"title": "新闻动态", "url": "/articles"}},
		"articles":         articles,
		"hot_articles":     []map[string]interface{}{},
		"categories":       categories,
		"current_category": currentCategory,
		"current_view":     view,
		"pagination":       pagination,
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/articles/list.html", ctx)
}

func (h *ArticleHandler) Detail(c echo.Context) error {
	base, err := LoadBaseData(c.Request().Context(), h.Settings, h.Friendly)
	if err != nil {
		return err
	}

	id := strings.TrimSpace(c.Param("id"))
	if id == "" {
		return kxlerrors.NotFound("not found: article not found")
	}

	a, err := h.Articles.GetPublic(c.Request().Context(), id)
	if err != nil {
		return err
	}

	// Category
	var category interface{} = nil
	if a.CategoryID != nil && h.DB != nil {
		var cat model.Category
		if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", *a.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
	}

	// Tags
	tagsByArticle, err := h.Articles.LoadTagsByArticleIDs(c.Request().Context(), []string{a.ID})
	if err != nil {
		return err
	}
	tags := []map[string]interface{}{}
	for _, t := range tagsByArticle[a.ID] {
		tags = append(tags, tagDTO(t))
	}

	article := map[string]interface{}{
		"id":           a.ID,
		"title":        a.Title,
		"summary":      a.Summary,
		"content":      a.Content,
		"cover_image":  a.CoverImage,
		"category":     category,
		"published_at": a.PublishedAt,
		"view_count":   a.ViewCount,
		"tags":         tags,
		"author":       nil,
		"created_at":   a.CreatedAt,
		"updated_at":   a.UpdatedAt,
	}

	// Prev/Next navigation.
	prevArticle := interface{}(nil)
	nextArticle := interface{}(nil)
	if prevID, nextID, err := h.Articles.NavigationPublic(c.Request().Context(), a.ID); err == nil {
		if prevID != nil {
			if pa, err := h.Articles.GetPublic(c.Request().Context(), *prevID); err == nil {
				prevArticle = map[string]interface{}{"id": pa.ID, "title": pa.Title}
			}
		}
		if nextID != nil {
			if na, err := h.Articles.GetPublic(c.Request().Context(), *nextID); err == nil {
				nextArticle = map[string]interface{}{"id": na.ID, "title": na.Title}
			}
		}
	}

	related := []map[string]interface{}{}
	if rows, err := h.Articles.RelatedPublic(c.Request().Context(), a.ID); err == nil {
		items, err := buildArticleListItems(c.Request().Context(), rows, h.Articles)
		if err == nil {
			related = items
		}
	}

	ctx := pongo2.Context{
		"page_title":      a.Title,
		"breadcrumbs":     []map[string]interface{}{{"title": "新闻动态", "url": "/articles"}, {"title": a.Title, "url": ""}},
		"article":         article,
		"prev_article":    prevArticle,
		"next_article":    nextArticle,
		"related_articles": related,
	}
	InjectBaseContext(ctx, c, base)
	return c.Render(http.StatusOK, "pages/articles/detail.html", ctx)
}

func categoryDTOs(rows []model.Category) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}
	out := make([]map[string]interface{}, 0, len(rows))
	for _, c := range rows {
		out = append(out, categoryDTO(c))
	}
	return out
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

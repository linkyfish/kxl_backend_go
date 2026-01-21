package admin

import (
	"context"
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

type ProjectHandler struct {
	DB       *gorm.DB
	Projects *service.ProjectService
}

func (h *ProjectHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:read"); err != nil {
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

	q := h.DB.WithContext(c.Request().Context()).Model(&model.Project{})
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(name ILIKE ? OR description ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	var rows []model.Project
	if err := q.Order("sort_order asc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	// Category & tags
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

type projectUpsertRequest struct {
	Name        string  `json:"name" form:"name"`
	Description string  `json:"description" form:"description"`
	CoverImage  *string `json:"cover_image" form:"cover_image"`
	CategoryID  *int    `json:"category_id" form:"category_id"`
	SortOrder   int     `json:"sort_order" form:"sort_order"`
}

func (h *ProjectHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	var req projectUpsertRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	p := &model.Project{
		Name:        req.Name,
		Description: req.Description,
		CoverImage:  normalizeOptString(req.CoverImage),
		CategoryID:  req.CategoryID,
		SortOrder:   req.SortOrder,
		Status:      0,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(p).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(h.projectDetailDTO(c.Request().Context(), *p)))
}

func (h *ProjectHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req projectUpsertRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var p model.Project
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}

	p.Name = req.Name
	p.Description = req.Description
	p.CoverImage = normalizeOptString(req.CoverImage)
	p.CategoryID = req.CategoryID
	p.SortOrder = req.SortOrder
	if err := h.DB.WithContext(c.Request().Context()).Save(&p).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(h.projectDetailDTO(c.Request().Context(), p)))
}

func (h *ProjectHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	id := c.Param("id")
	res := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).Delete(&model.Project{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: project not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type updateProjectStatusRequest struct {
	Status *int16 `json:"status" form:"status"`
}

func (h *ProjectHandler) UpdateStatus(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req updateProjectStatusRequest
	_ = c.Bind(&req)
	if req.Status == nil {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var p model.Project
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&p).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}
	p.Status = *req.Status
	if err := h.DB.WithContext(c.Request().Context()).Save(&p).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(h.projectDetailDTO(c.Request().Context(), p)))
}

func (h *ProjectHandler) ListFeatures(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:read"); err != nil {
		return err
	}
	id := c.Param("id")
	items, err := h.Projects.ListFeatures(c.Request().Context(), id)
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, f := range items {
		data = append(data, map[string]interface{}{
			"id":          f.ID,
			"name":        f.Name,
			"description": f.Description,
			"icon":        f.Icon,
			"sort_order":  f.SortOrder,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type featureUpsertRequest struct {
	Name        string  `json:"name" form:"name"`
	Description string  `json:"description" form:"description"`
	Icon        *string `json:"icon" form:"icon"`
	SortOrder   int     `json:"sort_order" form:"sort_order"`
}

func (h *ProjectHandler) CreateFeature(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	var req featureUpsertRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row := &model.ProjectFeature{
		ProjectID:   projectID,
		Name:        req.Name,
		Description: req.Description,
		Icon:        normalizeOptString(req.Icon),
		SortOrder:   req.SortOrder,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          row.ID,
		"name":        row.Name,
		"description": row.Description,
		"icon":        row.Icon,
		"sort_order":  row.SortOrder,
	}))
}

func (h *ProjectHandler) UpdateFeature(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	featureID, _ := strconv.Atoi(c.Param("featureId"))
	var req featureUpsertRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var row model.ProjectFeature
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", featureID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}
	row.ProjectID = projectID
	row.Name = req.Name
	row.Description = req.Description
	row.Icon = normalizeOptString(req.Icon)
	row.SortOrder = req.SortOrder
	if err := h.DB.WithContext(c.Request().Context()).Save(&row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":          row.ID,
		"name":        row.Name,
		"description": row.Description,
		"icon":        row.Icon,
		"sort_order":  row.SortOrder,
	}))
}

func (h *ProjectHandler) DeleteFeature(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	featureID, _ := strconv.Atoi(c.Param("featureId"))
	res := h.DB.WithContext(c.Request().Context()).Where("id = ?", featureID).Delete(&model.ProjectFeature{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: feature not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *ProjectHandler) ListMedia(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:read"); err != nil {
		return err
	}
	id := c.Param("id")
	items, err := h.Projects.ListMedia(c.Request().Context(), id)
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, m := range items {
		data = append(data, map[string]interface{}{
			"id":         m.ID,
			"type":       m.Type,
			"url":        m.URL,
			"title":      m.Title,
			"sort_order": m.SortOrder,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type mediaUpsertRequest struct {
	Type      string  `json:"type" form:"type"`
	URL       string  `json:"url" form:"url"`
	Title     *string `json:"title" form:"title"`
	SortOrder int     `json:"sort_order" form:"sort_order"`
}

func (h *ProjectHandler) CreateMedia(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	var req mediaUpsertRequest
	_ = c.Bind(&req)
	if req.Type == "" || req.URL == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	row := &model.ProjectMedia{
		ProjectID: projectID,
		Type:      req.Type,
		URL:       req.URL,
		Title:     normalizeOptString(req.Title),
		SortOrder: req.SortOrder,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"type":       row.Type,
		"url":        row.URL,
		"title":      row.Title,
		"sort_order": row.SortOrder,
	}))
}

func (h *ProjectHandler) UpdateMedia(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	mediaID, _ := strconv.Atoi(c.Param("mediaId"))
	var req mediaUpsertRequest
	_ = c.Bind(&req)
	if req.Type == "" || req.URL == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	var row model.ProjectMedia
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", mediaID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}
	row.ProjectID = projectID
	row.Type = req.Type
	row.URL = req.URL
	row.Title = normalizeOptString(req.Title)
	row.SortOrder = req.SortOrder
	if err := h.DB.WithContext(c.Request().Context()).Save(&row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         row.ID,
		"type":       row.Type,
		"url":        row.URL,
		"title":      row.Title,
		"sort_order": row.SortOrder,
	}))
}

func (h *ProjectHandler) DeleteMedia(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	mediaID, _ := strconv.Atoi(c.Param("mediaId"))
	res := h.DB.WithContext(c.Request().Context()).Where("id = ?", mediaID).Delete(&model.ProjectMedia{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: media not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *ProjectHandler) ListVersions(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:read"); err != nil {
		return err
	}
	projectID := c.Param("id")
	items, err := h.Projects.ListVersions(c.Request().Context(), projectID)
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, v := range items {
		date := ""
		if !v.ReleaseDate.IsZero() {
			date = v.ReleaseDate.UTC().Format("2006-01-02")
		}
		data = append(data, map[string]interface{}{
			"id":           v.ID,
			"version":      v.Version,
			"release_date": date,
			"changelog":    v.Changelog,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type versionUpsertRequest struct {
	Version     string `json:"version" form:"version"`
	ReleaseDate string `json:"release_date" form:"release_date"`
	Changelog   string `json:"changelog" form:"changelog"`
}

func (h *ProjectHandler) CreateVersion(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	var req versionUpsertRequest
	_ = c.Bind(&req)
	if req.Version == "" || req.ReleaseDate == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	rd, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return kxlerrors.Validation("validation error: invalid release_date")
	}
	row := &model.ProjectVersion{
		ProjectID:   projectID,
		Version:     req.Version,
		ReleaseDate: rd,
		Changelog:   req.Changelog,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":           row.ID,
		"version":      row.Version,
		"release_date": row.ReleaseDate.UTC().Format("2006-01-02"),
		"changelog":    row.Changelog,
	}))
}

func (h *ProjectHandler) UpdateVersion(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	versionID, _ := strconv.Atoi(c.Param("versionId"))
	var req versionUpsertRequest
	_ = c.Bind(&req)
	if req.Version == "" || req.ReleaseDate == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	rd, err := time.Parse("2006-01-02", req.ReleaseDate)
	if err != nil {
		return kxlerrors.Validation("validation error: invalid release_date")
	}
	var row model.ProjectVersion
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", versionID).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}
	row.ProjectID = projectID
	row.Version = req.Version
	row.ReleaseDate = rd
	row.Changelog = req.Changelog
	if err := h.DB.WithContext(c.Request().Context()).Save(&row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":           row.ID,
		"version":      row.Version,
		"release_date": row.ReleaseDate.UTC().Format("2006-01-02"),
		"changelog":    row.Changelog,
	}))
}

func (h *ProjectHandler) DeleteVersion(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	versionID, _ := strconv.Atoi(c.Param("versionId"))
	res := h.DB.WithContext(c.Request().Context()).Where("id = ?", versionID).Delete(&model.ProjectVersion{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: version not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type setTagsRequest struct {
	TagIDs []int `json:"tag_ids" form:"tag_ids"`
}

func (h *ProjectHandler) SetTags(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "projects:write"); err != nil {
		return err
	}
	projectID := c.Param("id")
	var req setTagsRequest
	_ = c.Bind(&req)
	if req.TagIDs == nil {
		req.TagIDs = []int{}
	}

	err := h.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("project_id = ?", projectID).Delete(&model.ProjectTag{}).Error; err != nil {
			return err
		}
		for _, tagID := range req.TagIDs {
			row := &model.ProjectTag{ProjectID: projectID, TagID: tagID}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return kxlerrors.Internal("db error")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *ProjectHandler) projectDetailDTO(ctx context.Context, p model.Project) map[string]interface{} {
	features, _ := h.Projects.ListFeatures(ctx, p.ID)
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

	media, _ := h.Projects.ListMedia(ctx, p.ID)
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

	versions, _ := h.Projects.ListVersions(ctx, p.ID)
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
		if err := h.DB.WithContext(ctx).Where("id = ?", *p.CategoryID).First(&cat).Error; err == nil {
			category = categoryDTO(cat)
		}
	}

	tagsByProject, _ := h.Projects.LoadTagsByProjectIDs(ctx, []string{p.ID})
	tags := []map[string]interface{}{}
	for _, t := range tagsByProject[p.ID] {
		tags = append(tags, tagDTO(t))
	}

	return map[string]interface{}{
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
	}
}

package admin

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type CaseHandler struct {
	DB       *gorm.DB
	Cases    *service.CaseService
	Projects *service.ProjectService
}

func (h *CaseHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "cases:read"); err != nil {
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

	q := h.DB.WithContext(c.Request().Context()).Model(&model.CaseStudy{})
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(client_name ILIKE ? OR summary ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	var rows []model.CaseStudy
	if err := q.Order("created_at desc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
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
	if err := middleware.AdminRequirePermission(c, "cases:read"); err != nil {
		return err
	}
	id := c.Param("id")

	var cs model.CaseStudy
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&cs).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: case not found")
		}
		return kxlerrors.Internal("db error")
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
			Order("sort_order asc").Order("id asc").
			Find(&projects).Error; err != nil {
			return kxlerrors.Internal("db error")
		}

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

type caseUpsertRequest struct {
	ClientName        string          `json:"client_name" form:"client_name"`
	CoverImage        *string         `json:"cover_image" form:"cover_image"`
	Summary           string          `json:"summary" form:"summary"`
	Background        string          `json:"background" form:"background"`
	Solution          string          `json:"solution" form:"solution"`
	Results           json.RawMessage `json:"results" form:"results"`
	Testimonial       *string         `json:"testimonial" form:"testimonial"`
	TestimonialAuthor *string         `json:"testimonial_author" form:"testimonial_author"`
	TestimonialTitle  *string         `json:"testimonial_title" form:"testimonial_title"`
	CategoryID        *int            `json:"category_id" form:"category_id"`
}

func (h *CaseHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "cases:write"); err != nil {
		return err
	}
	var req caseUpsertRequest
	_ = c.Bind(&req)
	if req.ClientName == "" || req.Summary == "" || req.Background == "" || req.Solution == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	results := datatypes.JSON([]byte("[]"))
	if len(req.Results) > 0 {
		if json.Valid(req.Results) {
			results = datatypes.JSON(req.Results)
		}
	}

	row := &model.CaseStudy{
		ClientName:        req.ClientName,
		CoverImage:        normalizeOptString(req.CoverImage),
		Summary:           req.Summary,
		Background:        req.Background,
		Solution:          req.Solution,
		Results:           results,
		Testimonial:       normalizeOptString(req.Testimonial),
		TestimonialAuthor: normalizeOptString(req.TestimonialAuthor),
		TestimonialTitle:  normalizeOptString(req.TestimonialTitle),
		CategoryID:        req.CategoryID,
		Status:            0,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":                 row.ID,
		"client_name":        row.ClientName,
		"cover_image":        row.CoverImage,
		"summary":            row.Summary,
		"background":         row.Background,
		"solution":           row.Solution,
		"results":            row.Results,
		"testimonial":        row.Testimonial,
		"testimonial_author": row.TestimonialAuthor,
		"testimonial_title":  row.TestimonialTitle,
		"status":             row.Status,
		"category":           nil,
		"related_projects":   []interface{}{},
		"created_at":         row.CreatedAt,
		"updated_at":         row.UpdatedAt,
	}))
}

func (h *CaseHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "cases:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req caseUpsertRequest
	_ = c.Bind(&req)
	if req.ClientName == "" || req.Summary == "" || req.Background == "" || req.Solution == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var row model.CaseStudy
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: case not found")
		}
		return kxlerrors.Internal("db error")
	}

	results := datatypes.JSON([]byte("[]"))
	if len(req.Results) > 0 {
		if json.Valid(req.Results) {
			results = datatypes.JSON(req.Results)
		}
	}

	row.ClientName = req.ClientName
	row.CoverImage = normalizeOptString(req.CoverImage)
	row.Summary = req.Summary
	row.Background = req.Background
	row.Solution = req.Solution
	row.Results = results
	row.Testimonial = normalizeOptString(req.Testimonial)
	row.TestimonialAuthor = normalizeOptString(req.TestimonialAuthor)
	row.TestimonialTitle = normalizeOptString(req.TestimonialTitle)
	row.CategoryID = req.CategoryID

	if err := h.DB.WithContext(c.Request().Context()).Save(&row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return h.Detail(c)
}

func (h *CaseHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "cases:write"); err != nil {
		return err
	}
	id := c.Param("id")
	res := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).Delete(&model.CaseStudy{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: case not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type caseStatusRequest struct {
	Status *int16 `json:"status" form:"status"`
}

func (h *CaseHandler) UpdateStatus(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "cases:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req caseStatusRequest
	_ = c.Bind(&req)
	if req.Status == nil {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var row model.CaseStudy
	if err := h.DB.WithContext(c.Request().Context()).Where("id = ?", id).First(&row).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: case not found")
		}
		return kxlerrors.Internal("db error")
	}
	row.Status = *req.Status
	if err := h.DB.WithContext(c.Request().Context()).Save(&row).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return h.Detail(c)
}

type caseProjectsRequest struct {
	ProjectIDs []string `json:"project_ids" form:"project_ids"`
}

func (h *CaseHandler) SetProjects(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "cases:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req caseProjectsRequest
	_ = c.Bind(&req)

	unique := make([]string, 0, len(req.ProjectIDs))
	seen := make(map[string]struct{})
	for _, pid := range req.ProjectIDs {
		if pid == "" {
			continue
		}
		if _, ok := seen[pid]; ok {
			continue
		}
		seen[pid] = struct{}{}
		unique = append(unique, pid)
	}

	if err := h.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("case_id = ?", id).Delete(&model.CaseProject{}).Error; err != nil {
			return kxlerrors.Internal("db error")
		}
		for _, pid := range unique {
			if err := tx.Create(&model.CaseProject{CaseID: id, ProjectID: pid}).Error; err != nil {
				return kxlerrors.Internal("db error")
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}


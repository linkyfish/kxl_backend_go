package admin

import (
	"net/http"
	"strconv"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type TestimonialHandler struct {
	Testimonials *service.TestimonialService
}

func (h *TestimonialHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.Testimonials.ListAll(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, t := range items {
		data = append(data, testimonialDTO(t))
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

func (h *TestimonialHandler) Detail(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	row, err := h.Testimonials.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(testimonialDTO(*row)))
}

type testimonialRequest struct {
	Name      string  `json:"name" form:"name"`
	Title     *string `json:"title" form:"title"`
	Company   *string `json:"company" form:"company"`
	Avatar    *string `json:"avatar" form:"avatar"`
	Content   string  `json:"content" form:"content"`
	Rating    int     `json:"rating" form:"rating"`
	SortOrder int     `json:"sort_order" form:"sort_order"`
	IsVisible *bool   `json:"is_visible" form:"is_visible"`
}

func (h *TestimonialHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req testimonialRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	rating := req.Rating
	if rating == 0 {
		rating = 5
	}
	row, err := h.Testimonials.Create(c.Request().Context(), &model.Testimonial{
		Name:      req.Name,
		Title:     normalizeOptString(req.Title),
		Company:   normalizeOptString(req.Company),
		Avatar:    normalizeOptString(req.Avatar),
		Content:   req.Content,
		Rating:    rating,
		SortOrder: req.SortOrder,
		IsVisible: isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(testimonialDTO(*row)))
}

func (h *TestimonialHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req testimonialRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	rating := req.Rating
	if rating == 0 {
		rating = 5
	}
	row, err := h.Testimonials.Update(c.Request().Context(), id, &model.Testimonial{
		Name:      req.Name,
		Title:     normalizeOptString(req.Title),
		Company:   normalizeOptString(req.Company),
		Avatar:    normalizeOptString(req.Avatar),
		Content:   req.Content,
		Rating:    rating,
		SortOrder: req.SortOrder,
		IsVisible: isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(testimonialDTO(*row)))
}

func (h *TestimonialHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	_ = h.Testimonials.Delete(c.Request().Context(), id)
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func testimonialDTO(t model.Testimonial) map[string]interface{} {
	return map[string]interface{}{
		"id":         t.ID,
		"name":       t.Name,
		"title":      t.Title,
		"company":    t.Company,
		"avatar":     t.Avatar,
		"content":    t.Content,
		"rating":     t.Rating,
		"sort_order": t.SortOrder,
		"is_visible": t.IsVisible,
		"created_at": t.CreatedAt,
		"updated_at": t.UpdatedAt,
	}
}


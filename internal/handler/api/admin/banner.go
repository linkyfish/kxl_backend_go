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

type BannerHandler struct {
	Banners *service.BannerService
}

func (h *BannerHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	items, err := h.Banners.ListAll(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(items))
	for _, b := range items {
		data = append(data, bannerDTO(b))
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

func (h *BannerHandler) Detail(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:read"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	b, err := h.Banners.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(bannerDTO(*b)))
}

type bannerRequest struct {
	Title     string  `json:"title" form:"title"`
	Subtitle  *string `json:"subtitle" form:"subtitle"`
	Highlight *string `json:"highlight" form:"highlight"`
	Tag       *string `json:"tag" form:"tag"`
	Image     *string `json:"image" form:"image"`
	Link      *string `json:"link" form:"link"`
	LinkText  string  `json:"link_text" form:"link_text"`
	BgClass   string  `json:"bg_class" form:"bg_class"`
	SortOrder int     `json:"sort_order" form:"sort_order"`
	IsVisible *bool   `json:"is_visible" form:"is_visible"`
}

func (h *BannerHandler) Create(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	var req bannerRequest
	_ = c.Bind(&req)
	if req.Title == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.Banners.Create(c.Request().Context(), &model.Banner{
		Title:     req.Title,
		Subtitle:  normalizeOptString(req.Subtitle),
		Highlight: normalizeOptString(req.Highlight),
		Tag:       normalizeOptString(req.Tag),
		Image:     normalizeOptString(req.Image),
		Link:      normalizeOptString(req.Link),
		LinkText:  req.LinkText,
		BgClass:   req.BgClass,
		SortOrder: req.SortOrder,
		IsVisible: isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(bannerDTO(*row)))
}

func (h *BannerHandler) Update(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req bannerRequest
	_ = c.Bind(&req)
	if req.Title == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	isVisible := true
	if req.IsVisible != nil {
		isVisible = *req.IsVisible
	}
	row, err := h.Banners.Update(c.Request().Context(), id, &model.Banner{
		Title:     req.Title,
		Subtitle:  normalizeOptString(req.Subtitle),
		Highlight: normalizeOptString(req.Highlight),
		Tag:       normalizeOptString(req.Tag),
		Image:     normalizeOptString(req.Image),
		Link:      normalizeOptString(req.Link),
		LinkText:  req.LinkText,
		BgClass:   req.BgClass,
		SortOrder: req.SortOrder,
		IsVisible: isVisible,
	})
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(bannerDTO(*row)))
}

func (h *BannerHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "settings:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	_ = h.Banners.Delete(c.Request().Context(), id)
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func bannerDTO(b model.Banner) map[string]interface{} {
	return map[string]interface{}{
		"id":         b.ID,
		"title":      b.Title,
		"subtitle":   b.Subtitle,
		"highlight":  b.Highlight,
		"tag":        b.Tag,
		"image":      b.Image,
		"link":       b.Link,
		"link_text":  b.LinkText,
		"bg_class":   b.BgClass,
		"sort_order": b.SortOrder,
		"is_visible": b.IsVisible,
		"created_at": b.CreatedAt,
		"updated_at": b.UpdatedAt,
	}
}


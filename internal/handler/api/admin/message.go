package admin

import (
	"net/http"
	"strconv"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	Messages *service.MessageService
}

func (h *MessageHandler) List(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:read"); err != nil {
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

	var statusPtr *int
	if raw := c.QueryParam("status"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			statusPtr = &n
		}
	}
	startDate := c.QueryParam("start_date")
	endDate := c.QueryParam("end_date")

	rows, total, err := h.Messages.List(c.Request().Context(), page, pageSize, statusPtr, startDate, endDate)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, m := range rows {
		items = append(items, map[string]interface{}{
			"id":         m.ID,
			"name":       m.Name,
			"company":    m.Company,
			"phone":      m.Phone,
			"email":      m.Email,
			"content":    m.Content,
			"status":     m.Status,
			"note":       m.Note,
			"created_at": m.CreatedAt,
			"updated_at": m.UpdatedAt,
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

func (h *MessageHandler) Detail(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:read"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	m, err := h.Messages.Get(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         m.ID,
		"name":       m.Name,
		"company":    m.Company,
		"phone":      m.Phone,
		"email":      m.Email,
		"content":    m.Content,
		"status":     m.Status,
		"note":       m.Note,
		"created_at": m.CreatedAt,
		"updated_at": m.UpdatedAt,
	}))
}

type updateMessageStatusRequest struct {
	Status *int16 `json:"status" form:"status"`
}

func (h *MessageHandler) UpdateStatus(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req updateMessageStatusRequest
	_ = c.Bind(&req)
	if req.Status == nil {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	m, err := h.Messages.UpdateStatus(c.Request().Context(), id, *req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         m.ID,
		"name":       m.Name,
		"company":    m.Company,
		"phone":      m.Phone,
		"email":      m.Email,
		"content":    m.Content,
		"status":     m.Status,
		"note":       m.Note,
		"created_at": m.CreatedAt,
		"updated_at": m.UpdatedAt,
	}))
}

type updateMessageNoteRequest struct {
	Note *string `json:"note" form:"note"`
}

func (h *MessageHandler) UpdateNote(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var req updateMessageNoteRequest
	_ = c.Bind(&req)
	if req.Note == nil {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	m, err := h.Messages.UpdateNote(c.Request().Context(), id, *req.Note)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         m.ID,
		"name":       m.Name,
		"company":    m.Company,
		"phone":      m.Phone,
		"email":      m.Email,
		"content":    m.Content,
		"status":     m.Status,
		"note":       m.Note,
		"created_at": m.CreatedAt,
		"updated_at": m.UpdatedAt,
	}))
}

func (h *MessageHandler) Delete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:write"); err != nil {
		return err
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Messages.Delete(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

type batchDeleteRequest struct {
	IDs []int `json:"ids" form:"ids"`
}

func (h *MessageHandler) BatchDelete(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:write"); err != nil {
		return err
	}
	var req batchDeleteRequest
	_ = c.Bind(&req)
	if req.IDs == nil {
		req.IDs = []int{}
	}
	if err := h.Messages.BatchDelete(c.Request().Context(), req.IDs); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *MessageHandler) Stats(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "messages:read"); err != nil {
		return err
	}
	stats, err := h.Messages.Stats(c.Request().Context())
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(stats))
}


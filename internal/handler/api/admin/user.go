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

type UserHandler struct {
	Users *service.UserService
}

func (h *UserHandler) ListUsers(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "users:read"); err != nil {
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

	keyword := c.QueryParam("keyword")
	var statusPtr *int
	if raw := c.QueryParam("status"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			statusPtr = &n
		}
	}

	rows, total, err := h.Users.ListUsers(c.Request().Context(), page, pageSize, keyword, statusPtr)
	if err != nil {
		return err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, u := range rows {
		items = append(items, map[string]interface{}{
			"id":         u.ID,
			"username":   u.Username,
			"email":      u.Email,
			"status":     u.Status,
			"created_at": u.CreatedAt,
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

func (h *UserHandler) DetailUser(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "users:read"); err != nil {
		return err
	}
	id := c.Param("id")
	u, err := h.Users.GetUser(c.Request().Context(), id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"status":     u.Status,
		"created_at": u.CreatedAt,
	}))
}

type updateUserStatusRequest struct {
	Status *int16 `json:"status" form:"status"`
}

func (h *UserHandler) UpdateUserStatus(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "users:write"); err != nil {
		return err
	}
	id := c.Param("id")
	var req updateUserStatusRequest
	_ = c.Bind(&req)
	if req.Status == nil {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	u, err := h.Users.UpdateUserStatus(c.Request().Context(), id, *req.Status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         u.ID,
		"username":   u.Username,
		"email":      u.Email,
		"status":     u.Status,
		"created_at": u.CreatedAt,
	}))
}

func (h *UserHandler) ListAdmins(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "admins:manage"); err != nil {
		return err
	}
	rows, err := h.Users.ListAdmins(c.Request().Context())
	if err != nil {
		return err
	}
	data := make([]map[string]interface{}, 0, len(rows))
	for _, a := range rows {
		data = append(data, map[string]interface{}{
			"id":         a.ID,
			"username":   a.Username,
			"role":       a.Role,
			"status":     a.Status,
			"created_at": a.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type createAdminRequest struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
	Role     string `json:"role" form:"role"`
	Status   *int16 `json:"status" form:"status"`
}

func (h *UserHandler) CreateAdmin(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "admins:manage"); err != nil {
		return err
	}
	var req createAdminRequest
	_ = c.Bind(&req)
	if req.Username == "" || req.Password == "" || req.Role == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	status := int16(1)
	if req.Status != nil {
		status = *req.Status
	}
	admin, err := h.Users.CreateAdmin(c.Request().Context(), req.Username, req.Password, req.Role, status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         admin.ID,
		"username":   admin.Username,
		"role":       admin.Role,
		"status":     admin.Status,
		"created_at": admin.CreatedAt,
	}))
}

type updateAdminRequest struct {
	Username string `json:"username" form:"username"`
	Role     string `json:"role" form:"role"`
	Status   *int16 `json:"status" form:"status"`
}

func (h *UserHandler) UpdateAdmin(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "admins:manage"); err != nil {
		return err
	}
	id := c.Param("id")
	var req updateAdminRequest
	_ = c.Bind(&req)
	if req.Username == "" || req.Role == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	status := int16(1)
	if req.Status != nil {
		status = *req.Status
	}
	admin, err := h.Users.UpdateAdmin(c.Request().Context(), id, req.Username, req.Role, status)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         admin.ID,
		"username":   admin.Username,
		"role":       admin.Role,
		"status":     admin.Status,
		"created_at": admin.CreatedAt,
	}))
}

func (h *UserHandler) DeleteAdmin(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "admins:manage"); err != nil {
		return err
	}
	id := c.Param("id")
	if err := h.Users.DeleteAdmin(c.Request().Context(), id); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}


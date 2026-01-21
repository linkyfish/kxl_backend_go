package admin

import (
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/middleware"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type RbacHandler struct {
	DB   *gorm.DB
	RBAC *service.RbacService
}

func (h *RbacHandler) ListRoles(c echo.Context) error {
	if !middleware.AdminHasPermission(c, "rbac:manage") && !middleware.AdminHasPermission(c, "admins:manage") {
		return kxlerrors.Forbidden()
	}

	var rows []model.AdminRole
	if err := h.DB.WithContext(c.Request().Context()).
		Order("is_system desc").
		Order("code asc").
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	data := make([]map[string]interface{}, 0, len(rows))
	for _, r := range rows {
		data = append(data, map[string]interface{}{
			"code":        r.Code,
			"name":        r.Name,
			"description": r.Description,
			"is_system":   r.IsSystem,
			"created_at":  r.CreatedAt,
			"updated_at":  r.UpdatedAt,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

type createRoleRequest struct {
	Code        string `json:"code" form:"code"`
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
}

func (h *RbacHandler) CreateRole(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "rbac:manage"); err != nil {
		return err
	}
	var req createRoleRequest
	_ = c.Bind(&req)
	if req.Code == "" || req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var count int64
	if err := h.DB.WithContext(c.Request().Context()).Model(&model.AdminRole{}).Where("code = ?", req.Code).Count(&count).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	if count > 0 {
		return kxlerrors.Conflict("conflict: role already exists")
	}

	role := &model.AdminRole{
		Code:        req.Code,
		Name:        req.Name,
		Description: req.Description,
		IsSystem:    false,
	}
	if err := h.DB.WithContext(c.Request().Context()).Create(role).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"code":        role.Code,
		"name":        role.Name,
		"description": role.Description,
		"is_system":   role.IsSystem,
		"created_at":  role.CreatedAt,
		"updated_at":  role.UpdatedAt,
	}))
}

type updateRoleRequest struct {
	Name        string `json:"name" form:"name"`
	Description string `json:"description" form:"description"`
}

func (h *RbacHandler) UpdateRole(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "rbac:manage"); err != nil {
		return err
	}
	code := c.Param("code")

	var req updateRoleRequest
	_ = c.Bind(&req)
	if req.Name == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}

	var role model.AdminRole
	if err := h.DB.WithContext(c.Request().Context()).Where("code = ?", code).First(&role).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return kxlerrors.NotFound("not found: resource not found")
		}
		return kxlerrors.Internal("db error")
	}

	role.Name = req.Name
	role.Description = req.Description
	if err := h.DB.WithContext(c.Request().Context()).Save(&role).Error; err != nil {
		return kxlerrors.Internal("db error")
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"code":        role.Code,
		"name":        role.Name,
		"description": role.Description,
		"is_system":   role.IsSystem,
		"created_at":  role.CreatedAt,
		"updated_at":  role.UpdatedAt,
	}))
}

func (h *RbacHandler) DeleteRole(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "rbac:manage"); err != nil {
		return err
	}
	code := c.Param("code")
	if code == "super_admin" || code == "admin" {
		return kxlerrors.Conflict("conflict: system role cannot be deleted")
	}

	var count int64
	_ = h.DB.WithContext(c.Request().Context()).Model(&model.Admin{}).Where("role = ?", code).Count(&count).Error
	if count > 0 {
		return kxlerrors.Conflict("conflict: role is in use")
	}

	res := h.DB.WithContext(c.Request().Context()).Where("code = ?", code).Delete(&model.AdminRole{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: role not found")
	}
	return c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (h *RbacHandler) ListPermissions(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "rbac:manage"); err != nil {
		return err
	}

	var rows []model.AdminPermission
	if err := h.DB.WithContext(c.Request().Context()).
		Order("group_name asc").
		Order("code asc").
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	data := make([]map[string]interface{}, 0, len(rows))
	for _, p := range rows {
		data = append(data, map[string]interface{}{
			"code":        p.Code,
			"name":        p.Name,
			"group_name":  p.GroupName,
			"description": p.Description,
			"is_system":   p.IsSystem,
		})
	}
	return c.JSON(http.StatusOK, response.Success(data))
}

func (h *RbacHandler) ListRolePermissions(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "rbac:manage"); err != nil {
		return err
	}
	code := c.Param("code")

	var count int64
	_ = h.DB.WithContext(c.Request().Context()).Model(&model.AdminRole{}).Where("code = ?", code).Count(&count).Error
	if count == 0 {
		return kxlerrors.NotFound("not found: role not found")
	}

	var rows []model.AdminRolePermission
	if err := h.DB.WithContext(c.Request().Context()).
		Where("role_code = ?", code).
		Order("permission_code asc").
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	perms := make([]string, 0, len(rows))
	for _, r := range rows {
		perms = append(perms, r.PermissionCode)
	}
	return c.JSON(http.StatusOK, response.Success(perms))
}

type setRolePermissionsRequest struct {
	PermissionCodes []string `json:"permission_codes" form:"permission_codes"`
}

func (h *RbacHandler) SetRolePermissions(c echo.Context) error {
	if err := middleware.AdminRequirePermission(c, "rbac:manage"); err != nil {
		return err
	}
	code := c.Param("code")
	if code == "super_admin" {
		return kxlerrors.Conflict("conflict: super_admin permissions are always allowed")
	}

	var roleCount int64
	_ = h.DB.WithContext(c.Request().Context()).Model(&model.AdminRole{}).Where("code = ?", code).Count(&roleCount).Error
	if roleCount == 0 {
		return kxlerrors.NotFound("not found: role not found")
	}

	var req setRolePermissionsRequest
	_ = c.Bind(&req)
	if req.PermissionCodes == nil {
		req.PermissionCodes = []string{}
	}

	// Validate permission codes exist.
	if len(req.PermissionCodes) > 0 {
		var valid []string
		if err := h.DB.WithContext(c.Request().Context()).
			Model(&model.AdminPermission{}).
			Where("code in ?", req.PermissionCodes).
			Pluck("code", &valid).Error; err != nil {
			return kxlerrors.Internal("db error")
		}
		if len(valid) != len(req.PermissionCodes) {
			return kxlerrors.Validation("validation error: invalid permission code")
		}
	}

	// Replace all permissions in a transaction.
	err := h.DB.WithContext(c.Request().Context()).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_code = ?", code).Delete(&model.AdminRolePermission{}).Error; err != nil {
			return err
		}
		for _, perm := range req.PermissionCodes {
			row := &model.AdminRolePermission{RoleCode: code, PermissionCode: perm}
			if err := tx.Create(row).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return kxlerrors.Internal("db error")
	}

	if h.RBAC != nil {
		h.RBAC.InvalidateRolePermissions(c.Request().Context(), code)
	}

	// Return updated permissions list.
	var rows []model.AdminRolePermission
	if err := h.DB.WithContext(c.Request().Context()).
		Where("role_code = ?", code).
		Order("permission_code asc").
		Find(&rows).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	perms := make([]string, 0, len(rows))
	for _, r := range rows {
		perms = append(perms, r.PermissionCode)
	}
	return c.JSON(http.StatusOK, response.Success(perms))
}


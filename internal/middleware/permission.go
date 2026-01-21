package middleware

import (
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

func AdminHasPermission(c echo.Context, code string) bool {
	role, _ := c.Get("current_admin_role").(string)
	perms, _ := c.Get("current_admin_permissions").([]string)
	return service.HasPermission(role, perms, code)
}

func AdminRequirePermission(c echo.Context, code string) error {
	if !AdminHasPermission(c, code) {
		return kxlerrors.Forbidden()
	}
	return nil
}


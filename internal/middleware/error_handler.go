package middleware

import (
	"errors"
	"net/http"
	"strings"

	kxlcfg "github.com/linkyfish/kxl_backend_go/internal/config"
	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/labstack/echo/v4"
)

func NewHTTPErrorHandler(cfg *kxlcfg.Config) echo.HTTPErrorHandler {
	return func(err error, c echo.Context) {
		if c.Response().Committed {
			return
		}

		path := ""
		wantsJSON := false
		if c.Request() != nil && c.Request().URL != nil {
			path = c.Request().URL.Path
			accept := c.Request().Header.Get("Accept")
			wantsJSON = strings.Contains(accept, "application/json")
		}
		isAPI := strings.HasPrefix(path, "/api/") || wantsJSON

		var be *kxlerrors.BusinessError
		if errors.As(err, &be) {
			if isAPI {
				_ = c.JSON(be.HTTPStatus, response.Error(be.Code, be.Message, be.Data))
				return
			}
			// SSR: render dedicated error pages when possible.
			if c.Echo().Renderer != nil && (be.HTTPStatus == http.StatusNotFound || be.HTTPStatus >= 500) {
				tpl := "pages/error/500.html"
				if be.HTTPStatus == http.StatusNotFound {
					tpl = "pages/error/404.html"
				}
				renderCtx := map[string]interface{}{
					"company":        map[string]interface{}{},
					"friendly_links": []map[string]interface{}{},
					"page_title":     "",
				}
				if rerr := c.Render(be.HTTPStatus, tpl, renderCtx); rerr == nil {
					return
				}
			}
			_ = c.String(be.HTTPStatus, be.Message)
			return
		}

		if he, ok := err.(*echo.HTTPError); ok {
			// Normalize to business error codes for API responses.
			switch he.Code {
			case http.StatusNotFound:
				be = kxlerrors.NotFound("not found")
			case http.StatusUnauthorized:
				be = kxlerrors.Unauthorized()
			case http.StatusForbidden:
				be = kxlerrors.Forbidden()
			case http.StatusTooManyRequests:
				be = kxlerrors.TooManyRequests()
			default:
				be = kxlerrors.New(kxlerrors.CodeValidationError, "bad request", he.Code, nil)
			}

			if isAPI {
				_ = c.JSON(be.HTTPStatus, response.Error(be.Code, be.Message, be.Data))
				return
			}
			if c.Echo().Renderer != nil && (be.HTTPStatus == http.StatusNotFound || be.HTTPStatus >= 500) {
				tpl := "pages/error/500.html"
				if be.HTTPStatus == http.StatusNotFound {
					tpl = "pages/error/404.html"
				}
				renderCtx := map[string]interface{}{
					"company":        map[string]interface{}{},
					"friendly_links": []map[string]interface{}{},
					"page_title":     "",
				}
				if rerr := c.Render(be.HTTPStatus, tpl, renderCtx); rerr == nil {
					return
				}
			}
			_ = c.String(be.HTTPStatus, be.Message)
			return
		}

		// Keep server error message minimal in production.
		msg := ""
		if cfg != nil && cfg.App.Env == "production" {
			msg = "internal error"
		} else {
			msg = err.Error()
		}

		be = kxlerrors.Internal(msg)
		if isAPI {
			_ = c.JSON(be.HTTPStatus, response.Error(be.Code, be.Message, be.Data))
			return
		}
		if c.Echo().Renderer != nil {
			renderCtx := map[string]interface{}{
				"company":        map[string]interface{}{},
				"friendly_links": []map[string]interface{}{},
				"page_title":     "",
			}
			if rerr := c.Render(be.HTTPStatus, "pages/error/500.html", renderCtx); rerr == nil {
				return
			}
		}
		_ = c.String(be.HTTPStatus, be.Message)
	}
}

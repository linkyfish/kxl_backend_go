package middleware

import (
	kxlcfg "github.com/linkyfish/kxl_backend_go/internal/config"
	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"
)

func CORS(cfg *kxlcfg.Config) echo.MiddlewareFunc {
	allowOrigin := "*"
	if cfg != nil && cfg.Cors.AllowOrigin != "" {
		allowOrigin = cfg.Cors.AllowOrigin
	}

	return echomw.CORSWithConfig(echomw.CORSConfig{
		AllowOrigins:     []string{allowOrigin},
		AllowMethods:     []string{echo.GET, echo.POST, echo.PUT, echo.PATCH, echo.DELETE, echo.OPTIONS},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		AllowCredentials: true,
	})
}


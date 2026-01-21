package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type HealthHandler struct {
	DB    *gorm.DB
	Redis *redis.Client
}

func (h *HealthHandler) Health(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

func (h *HealthHandler) Ready(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), 2*time.Second)
	defer cancel()

	dbOK := false
	if h.DB != nil {
		if sqlDB, err := h.DB.DB(); err == nil {
			dbOK = sqlDB.PingContext(ctx) == nil
		}
	}

	redisOK := false
	if h.Redis != nil {
		redisOK = h.Redis.Ping(ctx).Err() == nil
	}

	if dbOK && redisOK {
		return c.String(http.StatusOK, "ok")
	}
	return c.String(http.StatusServiceUnavailable, "not ready")
}


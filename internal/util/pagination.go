package util

import (
	"strconv"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/labstack/echo/v4"
)

type Pagination struct {
	Page     int64
	PageSize int64
}

func ParsePagination(c echo.Context) (*Pagination, error) {
	page := int64(1)
	pageSize := int64(10)

	if raw := c.QueryParam("page"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil && n > 0 {
			page = n
		}
	}
	if raw := c.QueryParam("page_size"); raw != "" {
		if n, err := strconv.ParseInt(raw, 10, 64); err == nil {
			pageSize = n
		}
	}
	if pageSize < 1 || pageSize > 200 {
		return nil, kxlerrors.Validation("validation error: page_size must be between 1 and 200")
	}

	return &Pagination{Page: page, PageSize: pageSize}, nil
}


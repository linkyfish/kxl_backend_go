package upload

import (
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type UploadHandler struct {
	Uploads *service.UploadService
}

func (h *UploadHandler) UploadImage(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil || f == nil {
		return kxlerrors.Validation("validation error: missing file")
	}
	url, err := h.Uploads.Upload(c.Request().Context(), f, service.UploadKindImage)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"url": url}))
}

func (h *UploadHandler) UploadVideo(c echo.Context) error {
	f, err := c.FormFile("file")
	if err != nil || f == nil {
		return kxlerrors.Validation("validation error: missing file")
	}
	url, err := h.Uploads.Upload(c.Request().Context(), f, service.UploadKindVideo)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{"url": url}))
}


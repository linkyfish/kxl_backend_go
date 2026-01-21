package v1

import (
	"net/http"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/dto/response"
	"github.com/linkyfish/kxl_backend_go/internal/service"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	Messages *service.MessageService
}

type submitMessageRequest struct {
	Name    string  `json:"name" form:"name"`
	Company *string `json:"company" form:"company"`
	Phone   string  `json:"phone" form:"phone"`
	Email   string  `json:"email" form:"email"`
	Content string  `json:"content" form:"content"`
}

func (h *MessageHandler) Submit(c echo.Context) error {
	var req submitMessageRequest
	_ = c.Bind(&req)
	if req.Name == "" || req.Phone == "" || req.Email == "" || req.Content == "" {
		return kxlerrors.Validation("validation error: missing required fields")
	}
	company := req.Company
	if company != nil && *company == "" {
		company = nil
	}

	msg, err := h.Messages.Submit(c.Request().Context(), req.Name, company, req.Phone, req.Email, req.Content)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response.Success(map[string]interface{}{
		"id":         msg.ID,
		"name":       msg.Name,
		"company":    msg.Company,
		"phone":      msg.Phone,
		"email":      msg.Email,
		"content":    msg.Content,
		"status":     msg.Status,
		"note":       msg.Note,
		"created_at": msg.CreatedAt,
		"updated_at": msg.UpdatedAt,
	}))
}


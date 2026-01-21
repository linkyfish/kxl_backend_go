package validator

import (
	"fmt"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	v := validator.New()
	return &Validator{validate: v}
}

func (v *Validator) Validate(i interface{}) error {
	if err := v.validate.Struct(i); err != nil {
		return kxlerrors.Validation(formatValidationError(err))
	}
	return nil
}

func formatValidationError(err error) string {
	if ve, ok := err.(validator.ValidationErrors); ok && len(ve) > 0 {
		// Keep message compact & stable; match "validation error: ..." style from PHP backend.
		fe := ve[0]
		return fmt.Sprintf("validation error: %s failed on %s", fe.Field(), fe.Tag())
	}
	return "validation error: invalid request"
}


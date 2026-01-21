package errors

import "net/http"

const (
	CodeValidationError = 40001
	CodeConflict        = 40002
	CodeUnauthorized    = 40101
	CodeForbidden       = 40301
	CodeNotFound        = 40401
	CodeTooManyRequests = 42901

	CodeInternalError = 50099
)

type BusinessError struct {
	Code       int
	Message    string
	HTTPStatus int
	Data       interface{}
}

func (e *BusinessError) Error() string {
	return e.Message
}

func New(code int, message string, httpStatus int, data interface{}) *BusinessError {
	return &BusinessError{Code: code, Message: message, HTTPStatus: httpStatus, Data: data}
}

func Validation(message string) *BusinessError {
	return New(CodeValidationError, message, http.StatusBadRequest, nil)
}

func Conflict(message string) *BusinessError {
	return New(CodeConflict, message, http.StatusConflict, nil)
}

func Unauthorized() *BusinessError {
	return New(CodeUnauthorized, "unauthorized", http.StatusUnauthorized, nil)
}

func Forbidden() *BusinessError {
	return New(CodeForbidden, "forbidden", http.StatusForbidden, nil)
}

func NotFound(message string) *BusinessError {
	if message == "" {
		message = "not found"
	}
	return New(CodeNotFound, message, http.StatusNotFound, nil)
}

func TooManyRequests() *BusinessError {
	return New(CodeTooManyRequests, "too many requests", http.StatusTooManyRequests, nil)
}

func Internal(message string) *BusinessError {
	if message == "" {
		message = "internal error"
	}
	return New(CodeInternalError, message, http.StatusInternalServerError, nil)
}

// HTTPStatusFromCode mirrors kxl_backend_php/support/helpers.php mapping.
func HTTPStatusFromCode(code int) int {
	switch {
	case code == CodeValidationError:
		return http.StatusBadRequest
	case code == CodeConflict:
		return http.StatusConflict
	case code == CodeUnauthorized:
		return http.StatusUnauthorized
	case code == CodeForbidden:
		return http.StatusForbidden
	case code == CodeNotFound:
		return http.StatusNotFound
	case code == CodeTooManyRequests:
		return http.StatusTooManyRequests
	case code >= 50000:
		return http.StatusInternalServerError
	default:
		return http.StatusBadRequest
	}
}


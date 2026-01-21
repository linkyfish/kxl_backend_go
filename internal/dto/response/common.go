package response

// ApiResponse matches Rust/PHP response envelope:
// { "code": 200, "message": "success", "data": ... }
type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func Success(data interface{}) ApiResponse {
	return ApiResponse{Code: 200, Message: "success", Data: data}
}

func SuccessWithoutData() ApiResponse {
	return ApiResponse{Code: 200, Message: "success", Data: nil}
}

func Error(code int, message string, data interface{}) ApiResponse {
	return ApiResponse{Code: code, Message: message, Data: data}
}

// Paged matches Rust kxl_backend_core::dto::response::Paged<T>.
type Paged struct {
	Items      interface{} `json:"items"`
	Total      int64       `json:"total"`
	Page       int64       `json:"page"`
	PageSize   int64       `json:"page_size"`
	TotalPages int64       `json:"total_pages"`
}

func Paginated(items interface{}, total, page, pageSize int64) Paged {
	var totalPages int64
	if total == 0 {
		totalPages = 0
	} else {
		totalPages = (total + pageSize - 1) / pageSize
	}
	return Paged{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
}


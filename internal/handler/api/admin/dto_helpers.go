package admin

import "github.com/linkyfish/kxl_backend_go/internal/model"

func categoryDTO(c model.Category) map[string]interface{} {
	return map[string]interface{}{
		"id":         c.ID,
		"name":       c.Name,
		"type":       c.Type,
		"sort_order": c.SortOrder,
		"created_at": c.CreatedAt,
	}
}

func tagDTO(t model.Tag) map[string]interface{} {
	return map[string]interface{}{
		"id":         t.ID,
		"name":       t.Name,
		"type":       t.Type,
		"created_at": t.CreatedAt,
	}
}


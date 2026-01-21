package web

import (
	"context"

	"github.com/linkyfish/kxl_backend_go/internal/model"
	"github.com/linkyfish/kxl_backend_go/internal/service"
)

func buildArticleListItems(ctx context.Context, rows []model.Article, svc *service.ArticleService) ([]map[string]interface{}, error) {
	if len(rows) == 0 {
		return []map[string]interface{}{}, nil
	}

	articleIDs := make([]string, 0, len(rows))
	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	for _, a := range rows {
		articleIDs = append(articleIDs, a.ID)
		if a.CategoryID != nil {
			if _, ok := seenCat[*a.CategoryID]; !ok {
				seenCat[*a.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *a.CategoryID)
			}
		}
	}

	categoryMap, err := svc.LoadCategoryMap(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}
	tagsByArticle, err := svc.LoadTagsByArticleIDs(ctx, articleIDs)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, a := range rows {
		var category interface{} = nil
		if a.CategoryID != nil {
			if cat, ok := categoryMap[*a.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}
		tags := []map[string]interface{}{}
		for _, t := range tagsByArticle[a.ID] {
			tags = append(tags, tagDTO(t))
		}
		items = append(items, map[string]interface{}{
			"id":           a.ID,
			"title":        a.Title,
			"summary":      a.Summary,
			"cover_image":  a.CoverImage,
			"category":     category,
			"published_at": a.PublishedAt,
			"view_count":   a.ViewCount,
			"tags":         tags,
			"created_at":   a.CreatedAt,
			"updated_at":   a.UpdatedAt,
		})
	}
	return items, nil
}

func buildProjectListItems(ctx context.Context, rows []model.Project, svc *service.ProjectService) ([]map[string]interface{}, error) {
	if len(rows) == 0 {
		return []map[string]interface{}{}, nil
	}

	projectIDs := make([]string, 0, len(rows))
	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	for _, p := range rows {
		projectIDs = append(projectIDs, p.ID)
		if p.CategoryID != nil {
			if _, ok := seenCat[*p.CategoryID]; !ok {
				seenCat[*p.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *p.CategoryID)
			}
		}
	}

	categoryMap, err := svc.LoadCategoryMap(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}
	tagsByProject, err := svc.LoadTagsByProjectIDs(ctx, projectIDs)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, p := range rows {
		var category interface{} = nil
		if p.CategoryID != nil {
			if cat, ok := categoryMap[*p.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}
		tags := []map[string]interface{}{}
		for _, t := range tagsByProject[p.ID] {
			tags = append(tags, tagDTO(t))
		}
		items = append(items, map[string]interface{}{
			"id":          p.ID,
			"name":        p.Name,
			"description": p.Description,
			"cover_image": p.CoverImage,
			"status":      p.Status,
			"sort_order":  p.SortOrder,
			"category":    category,
			"tags":        tags,
			"created_at":  p.CreatedAt,
			"updated_at":  p.UpdatedAt,

			// Optional fields used by templates.
			"platform":   nil,
			"tech_stack": nil,
		})
	}
	return items, nil
}

func buildCaseListItems(ctx context.Context, rows []model.CaseStudy, svc *service.CaseService) ([]map[string]interface{}, error) {
	if len(rows) == 0 {
		return []map[string]interface{}{}, nil
	}

	categoryIDs := make([]int, 0)
	seenCat := make(map[int]struct{})
	for _, row := range rows {
		if row.CategoryID != nil {
			if _, ok := seenCat[*row.CategoryID]; !ok {
				seenCat[*row.CategoryID] = struct{}{}
				categoryIDs = append(categoryIDs, *row.CategoryID)
			}
		}
	}
	categoryMap, err := svc.LoadCategoryMap(ctx, categoryIDs)
	if err != nil {
		return nil, err
	}

	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		var category interface{} = nil
		if row.CategoryID != nil {
			if cat, ok := categoryMap[*row.CategoryID]; ok {
				category = categoryDTO(cat)
			}
		}
		items = append(items, map[string]interface{}{
			"id":               row.ID,
			"client_name":      row.ClientName,
			"client_logo":      nil,
			"cover_image":      row.CoverImage,
			"summary":          row.Summary,
			"status":           row.Status,
			"category":         category,
			"related_projects": []map[string]interface{}{},
		})
	}
	return items, nil
}


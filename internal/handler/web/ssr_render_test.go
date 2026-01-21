package web

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/flosch/pongo2/v6"
)

func TestSSRPagesRenderWithMinimalContext(t *testing.T) {
	templateDir := filepath.Join("..", "..", "..", "templates")
	r, err := NewRenderer(templateDir)
	if err != nil {
		t.Fatalf("NewRenderer: %v", err)
	}

	base := pongo2.Context{
		"company":        companyDTO(nil),
		"friendly_links": []map[string]interface{}{},
		"current_path":   "/",
	}

	cases := []struct {
		name string
		ctx  pongo2.Context
	}{
		{"pages/home.html", mergeCtx(base, pongo2.Context{
			"banners":           []map[string]interface{}{},
			"featured_projects": []map[string]interface{}{},
			"featured_cases":    []map[string]interface{}{},
			"latest_articles":   []map[string]interface{}{},
			"testimonials":      []map[string]interface{}{},
			"solutions":         []map[string]interface{}{},
			"partners":          []map[string]interface{}{},
		})},
		{"pages/about.html", mergeCtx(base, pongo2.Context{
			"milestones":   []map[string]interface{}{},
			"team_members": []map[string]interface{}{},
		})},
		{"pages/contact.html", base},
		{"pages/search.html", mergeCtx(base, pongo2.Context{
			"keyword":     "",
			"search_type": "",
			"results":     []map[string]interface{}{},
			"total":       0,
			"pagination": map[string]interface{}{
				"current_page": 1,
				"total_pages":  0,
				"total_items":  0,
				"base_url":     "/search",
				"query":        "",
			},
		})},
		{"pages/login.html", base},
		{"pages/register.html", base},

		{"pages/articles/list.html", mergeCtx(base, pongo2.Context{
			"articles":         []map[string]interface{}{},
			"hot_articles":     []map[string]interface{}{},
			"categories":       []map[string]interface{}{},
			"current_category": nil,
			"current_view":     "grid",
			"pagination": map[string]interface{}{
				"current_page": 1,
				"total_pages":  0,
				"total_items":  0,
				"base_url":     "/articles",
				"query":        "",
			},
		})},
		{"pages/articles/detail.html", mergeCtx(base, pongo2.Context{
			"article": map[string]interface{}{
				"id":           "test",
				"title":        "Test Article",
				"summary":      "",
				"content":      "",
				"cover_image":  nil,
				"category":     nil,
				"published_at": "",
				"view_count":   0,
				"tags":         []map[string]interface{}{},
				"created_at":   "",
			},
			"prev_article":     nil,
			"next_article":     nil,
			"related_articles": []map[string]interface{}{},
		})},

		{"pages/projects/list.html", mergeCtx(base, pongo2.Context{
			"projects":         []map[string]interface{}{},
			"categories":       []map[string]interface{}{},
			"current_category": nil,
			"pagination": map[string]interface{}{
				"current_page": 1,
				"total_pages":  0,
				"total_items":  0,
				"base_url":     "/projects",
				"query":        "",
			},
		})},
		{"pages/projects/detail.html", mergeCtx(base, pongo2.Context{
			"project": map[string]interface{}{
				"id":          "test",
				"name":        "Test Project",
				"description": "",
				"cover_image": nil,
				"category":    nil,
				"tags":        []map[string]interface{}{},
				"features":    []map[string]interface{}{},
				"media":       []map[string]interface{}{},
				"versions":    []map[string]interface{}{},
				"created_at":  "",
			},
		})},

		{"pages/cases/list.html", mergeCtx(base, pongo2.Context{
			"cases":            []map[string]interface{}{},
			"categories":       []map[string]interface{}{},
			"current_category": nil,
			"pagination": map[string]interface{}{
				"current_page": 1,
				"total_pages":  0,
				"total_items":  0,
				"base_url":     "/cases",
				"query":        "",
			},
		})},
		{"pages/cases/detail.html", mergeCtx(base, pongo2.Context{
			"case": map[string]interface{}{
				"id":               "test",
				"client_name":      "Test Client",
				"summary":          "",
				"cover_image":      nil,
				"results":          []interface{}{},
				"background":       "",
				"solution":         "",
				"testimonial":      nil,
				"category":         nil,
				"related_projects": []map[string]interface{}{},
			},
		})},

		{"pages/error/404.html", base},
		{"pages/error/500.html", base},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tpl, err := r.get(tc.name)
			if err != nil {
				t.Fatalf("compile %s: %v", tc.name, err)
			}
			var buf bytes.Buffer
			if err := tpl.ExecuteWriter(tc.ctx, &buf); err != nil {
				t.Fatalf("render %s: %v", tc.name, err)
			}
		})
	}
}

func mergeCtx(base pongo2.Context, extra pongo2.Context) pongo2.Context {
	out := pongo2.Context{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}


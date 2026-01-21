package service

import (
	"context"
	"regexp"
	"sort"
	"strings"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type SearchService struct {
	db *gorm.DB
}

func NewSearchService(db *gorm.DB) *SearchService {
	return &SearchService{db: db}
}

type SearchResultItem struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
	Highlight string `json:"highlight"`
	URL       string `json:"url"`
}

func (s *SearchService) Search(ctx context.Context, q string, typ *string, page, pageSize int64) ([]SearchResultItem, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, kxlerrors.Internal("db not configured")
	}
	q = strings.TrimSpace(q)
	if q == "" {
		return nil, 0, kxlerrors.Validation("validation error: q is required")
	}

	highlight := makeHighlighter(q)

	offset := (page - 1) * pageSize
	fetchLimit := page * pageSize

	if typ != nil {
		switch *typ {
		case "project":
			items, total, err := s.searchProjects(ctx, q, pageSize, offset)
			if err != nil {
				return nil, 0, err
			}
			out := make([]SearchResultItem, 0, len(items))
			for _, p := range items {
				out = append(out, SearchResultItem{
					ID:        p.ID,
					Type:      "project",
					Title:     p.Name,
					Summary:   p.Description,
					Highlight: highlight(p.Description),
					URL:       "/projects/" + p.ID,
				})
			}
			return out, total, nil
		case "article":
			items, total, err := s.searchArticles(ctx, q, pageSize, offset)
			if err != nil {
				return nil, 0, err
			}
			out := make([]SearchResultItem, 0, len(items))
			for _, a := range items {
				out = append(out, SearchResultItem{
					ID:        a.ID,
					Type:      "article",
					Title:     a.Title,
					Summary:   a.Summary,
					Highlight: highlight(a.Summary),
					URL:       "/articles/" + a.ID,
				})
			}
			return out, total, nil
		case "case":
			items, total, err := s.searchCases(ctx, q, pageSize, offset)
			if err != nil {
				return nil, 0, err
			}
			out := make([]SearchResultItem, 0, len(items))
			for _, c := range items {
				out = append(out, SearchResultItem{
					ID:        c.ID,
					Type:      "case",
					Title:     c.ClientName,
					Summary:   c.Summary,
					Highlight: highlight(c.Summary),
					URL:       "/cases/" + c.ID,
				})
			}
			return out, total, nil
		}
	}

	projects, t1, err := s.searchProjects(ctx, q, fetchLimit, 0)
	if err != nil {
		return nil, 0, err
	}
	articles, t2, err := s.searchArticles(ctx, q, fetchLimit, 0)
	if err != nil {
		return nil, 0, err
	}
	cases, t3, err := s.searchCases(ctx, q, fetchLimit, 0)
	if err != nil {
		return nil, 0, err
	}
	total := t1 + t2 + t3

	type hit struct {
		ts   int64
		item SearchResultItem
	}
	hits := make([]hit, 0, len(projects)+len(articles)+len(cases))
	for _, p := range projects {
		hits = append(hits, hit{
			ts: p.CreatedAt.Unix(),
			item: SearchResultItem{
				ID:        p.ID,
				Type:      "project",
				Title:     p.Name,
				Summary:   p.Description,
				Highlight: highlight(p.Description),
				URL:       "/projects/" + p.ID,
			},
		})
	}
	for _, a := range articles {
		hits = append(hits, hit{
			ts: a.CreatedAt.Unix(),
			item: SearchResultItem{
				ID:        a.ID,
				Type:      "article",
				Title:     a.Title,
				Summary:   a.Summary,
				Highlight: highlight(a.Summary),
				URL:       "/articles/" + a.ID,
			},
		})
	}
	for _, c := range cases {
		hits = append(hits, hit{
			ts: c.CreatedAt.Unix(),
			item: SearchResultItem{
				ID:        c.ID,
				Type:      "case",
				Title:     c.ClientName,
				Summary:   c.Summary,
				Highlight: highlight(c.Summary),
				URL:       "/cases/" + c.ID,
			},
		})
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].ts > hits[j].ts })

	start := int(offset)
	if start > len(hits) {
		start = len(hits)
	}
	end := start + int(pageSize)
	if end > len(hits) {
		end = len(hits)
	}

	out := make([]SearchResultItem, 0, end-start)
	for _, h := range hits[start:end] {
		out = append(out, h.item)
	}
	return out, total, nil
}

func (s *SearchService) Suggestions(ctx context.Context) ([]string, error) {
	// Keep behavior identical to PHP backend for now.
	return []string{}, nil
}

func (s *SearchService) searchProjects(ctx context.Context, q string, limit, offset int64) ([]model.Project, int64, error) {
	pattern := "%" + q + "%"
	base := s.db.WithContext(ctx).Model(&model.Project{}).
		Where("status = ?", 1).
		Where("(name ILIKE ? OR description ILIKE ?)", pattern, pattern)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	var rows []model.Project
	if err := base.Order("created_at desc").Limit(int(limit)).Offset(int(offset)).Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *SearchService) searchArticles(ctx context.Context, q string, limit, offset int64) ([]model.Article, int64, error) {
	pattern := "%" + q + "%"
	base := s.db.WithContext(ctx).Model(&model.Article{}).
		Where("status = ?", 1).
		Where("(title ILIKE ? OR summary ILIKE ? OR content ILIKE ?)", pattern, pattern, pattern)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	var rows []model.Article
	if err := base.Order("created_at desc").Limit(int(limit)).Offset(int(offset)).Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *SearchService) searchCases(ctx context.Context, q string, limit, offset int64) ([]model.CaseStudy, int64, error) {
	pattern := "%" + q + "%"
	base := s.db.WithContext(ctx).Model(&model.CaseStudy{}).
		Where("status = ?", 1).
		Where("(client_name ILIKE ? OR summary ILIKE ? OR background ILIKE ? OR solution ILIKE ?)", pattern, pattern, pattern, pattern)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	var rows []model.CaseStudy
	if err := base.Order("created_at desc").Limit(int(limit)).Offset(int(offset)).Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func makeHighlighter(q string) func(string) string {
	escaped := regexp.QuoteMeta(q)
	re := regexp.MustCompile("(?i)" + escaped)
	return func(text string) string {
		if text == "" {
			return ""
		}
		return re.ReplaceAllString(text, "<em>$0</em>")
	}
}


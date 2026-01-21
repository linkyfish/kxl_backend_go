package service

import (
	"context"
	"errors"
	"time"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type ArticleService struct {
	db *gorm.DB
}

func NewArticleService(db *gorm.DB) *ArticleService {
	return &ArticleService{db: db}
}

func (s *ArticleService) ListPublic(ctx context.Context, page, pageSize int64, categoryID *int, keyword string) ([]model.Article, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.Article{}).Where("status = ?", 1)
	if categoryID != nil {
		q = q.Where("category_id = ?", *categoryID)
	}
	if keyword != "" {
		pattern := "%" + keyword + "%"
		q = q.Where("(title ILIKE ? OR summary ILIKE ?)", pattern, pattern)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}

	var rows []model.Article
	if err := q.Order("published_at desc").Order("id asc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *ArticleService) GetPublic(ctx context.Context, id string) (*model.Article, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var a model.Article
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&a).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, kxlerrors.NotFound("not found: article not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	if a.Status != 1 {
		return nil, kxlerrors.NotFound("not found: article not found")
	}
	return &a, nil
}

func (s *ArticleService) IncrementViewCount(ctx context.Context, id string) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	return s.db.WithContext(ctx).
		Model(&model.Article{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + 1")).Error
}

func (s *ArticleService) RelatedPublic(ctx context.Context, id string) ([]model.Article, error) {
	a, err := s.GetPublic(ctx, id)
	if err != nil {
		return nil, err
	}

	q := s.db.WithContext(ctx).Model(&model.Article{}).
		Where("status = ?", 1).
		Where("id <> ?", a.ID)
	if a.CategoryID != nil {
		q = q.Where("category_id = ?", *a.CategoryID)
	}
	var rows []model.Article
	if err := q.Order("published_at desc").Order("id asc").Limit(5).Find(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return rows, nil
}

func (s *ArticleService) NavigationPublic(ctx context.Context, id string) (*string, *string, error) {
	a, err := s.GetPublic(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if a.PublishedAt == nil {
		return nil, nil, nil
	}
	publishedAt := *a.PublishedAt

	var prevID *string
	{
		var row struct {
			ID string `gorm:"column:id"`
		}
		err := s.db.WithContext(ctx).
			Table("articles").
			Select("id").
			Where("status = ? AND published_at < ?", 1, publishedAt).
			Order("published_at desc").
			Limit(1).
			Take(&row).Error
		if err == nil {
			prevID = &row.ID
		}
	}

	var nextID *string
	{
		var row struct {
			ID string `gorm:"column:id"`
		}
		err := s.db.WithContext(ctx).
			Table("articles").
			Select("id").
			Where("status = ? AND published_at > ?", 1, publishedAt).
			Order("published_at asc").
			Limit(1).
			Take(&row).Error
		if err == nil {
			nextID = &row.ID
		}
	}

	return prevID, nextID, nil
}

func (s *ArticleService) LoadCategoryMap(ctx context.Context, categoryIDs []int) (map[int]model.Category, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if len(categoryIDs) == 0 {
		return map[int]model.Category{}, nil
	}
	var cats []model.Category
	if err := s.db.WithContext(ctx).Where("id in ?", categoryIDs).Find(&cats).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	m := make(map[int]model.Category, len(cats))
	for _, c := range cats {
		m[c.ID] = c
	}
	return m, nil
}

func (s *ArticleService) LoadTagsByArticleIDs(ctx context.Context, articleIDs []string) (map[string][]model.Tag, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	if len(articleIDs) == 0 {
		return map[string][]model.Tag{}, nil
	}

	type row struct {
		ArticleID string    `gorm:"column:article_id"`
		ID        int       `gorm:"column:id"`
		Name      string    `gorm:"column:name"`
		Type      string    `gorm:"column:type"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	if err := s.db.WithContext(ctx).
		Table("article_tags").
		Select("article_tags.article_id as article_id, tags.id as id, tags.name as name, tags.type as type, tags.created_at as created_at").
		Joins("join tags on article_tags.tag_id = tags.id").
		Where("article_tags.article_id in ?", articleIDs).
		Order("article_tags.article_id asc").
		Order("tags.id asc").
		Scan(&rows).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}

	out := make(map[string][]model.Tag)
	for _, r := range rows {
		out[r.ArticleID] = append(out[r.ArticleID], model.Tag{
			ID:        r.ID,
			Name:      r.Name,
			Type:      r.Type,
			CreatedAt: r.CreatedAt,
		})
	}
	return out, nil
}


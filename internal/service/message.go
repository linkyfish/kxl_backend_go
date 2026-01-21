package service

import (
	"context"
	"time"

	kxlerrors "github.com/linkyfish/kxl_backend_go/internal/errors"
	"github.com/linkyfish/kxl_backend_go/internal/model"
	"gorm.io/gorm"
)

type MessageService struct {
	db *gorm.DB
}

func NewMessageService(db *gorm.DB) *MessageService {
	return &MessageService{db: db}
}

func (s *MessageService) Submit(ctx context.Context, name string, company *string, phone string, email string, content string) (*model.Message, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	msg := &model.Message{
		Name:    name,
		Company: company,
		Phone:   phone,
		Email:   email,
		Content: content,
		Status:  0,
		Note:    nil,
	}
	if err := s.db.WithContext(ctx).Create(msg).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return msg, nil
}

func (s *MessageService) List(ctx context.Context, page, pageSize int64, status *int, startDate, endDate string) ([]model.Message, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, kxlerrors.Internal("db not configured")
	}
	q := s.db.WithContext(ctx).Model(&model.Message{})
	if status != nil {
		q = q.Where("status = ?", *status)
	}
	if startDate != "" {
		if t, err := time.Parse("2006-01-02", startDate); err == nil {
			q = q.Where("created_at >= ?", t.UTC())
		}
	}
	if endDate != "" {
		if t, err := time.Parse("2006-01-02", endDate); err == nil {
			end := time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
			q = q.Where("created_at <= ?", end)
		}
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}

	var rows []model.Message
	if err := q.Order("created_at desc").Order("id desc").
		Offset(int((page - 1) * pageSize)).Limit(int(pageSize)).
		Find(&rows).Error; err != nil {
		return nil, 0, kxlerrors.Internal("db error")
	}
	return rows, total, nil
}

func (s *MessageService) Get(ctx context.Context, id int) (*model.Message, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var msg model.Message
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&msg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	return &msg, nil
}

func (s *MessageService) UpdateStatus(ctx context.Context, id int, status int16) (*model.Message, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var msg model.Message
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&msg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	msg.Status = status
	if err := s.db.WithContext(ctx).Save(&msg).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &msg, nil
}

func (s *MessageService) UpdateNote(ctx context.Context, id int, note string) (*model.Message, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var msg model.Message
	if err := s.db.WithContext(ctx).Where("id = ?", id).First(&msg).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, kxlerrors.NotFound("not found: resource not found")
		}
		return nil, kxlerrors.Internal("db error")
	}
	msg.Note = &note
	if err := s.db.WithContext(ctx).Save(&msg).Error; err != nil {
		return nil, kxlerrors.Internal("db error")
	}
	return &msg, nil
}

func (s *MessageService) Delete(ctx context.Context, id int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	res := s.db.WithContext(ctx).Where("id = ?", id).Delete(&model.Message{})
	if res.Error != nil {
		return kxlerrors.Internal("db error")
	}
	if res.RowsAffected == 0 {
		return kxlerrors.NotFound("not found: message not found")
	}
	return nil
}

func (s *MessageService) BatchDelete(ctx context.Context, ids []int) error {
	if s == nil || s.db == nil {
		return kxlerrors.Internal("db not configured")
	}
	if len(ids) == 0 {
		return nil
	}
	if err := s.db.WithContext(ctx).Where("id in ?", ids).Delete(&model.Message{}).Error; err != nil {
		return kxlerrors.Internal("db error")
	}
	return nil
}

func (s *MessageService) Stats(ctx context.Context) (map[string]int64, error) {
	if s == nil || s.db == nil {
		return nil, kxlerrors.Internal("db not configured")
	}
	var unprocessed, processed, replied int64
	_ = s.db.WithContext(ctx).Model(&model.Message{}).Where("status = ?", 0).Count(&unprocessed).Error
	_ = s.db.WithContext(ctx).Model(&model.Message{}).Where("status = ?", 1).Count(&processed).Error
	_ = s.db.WithContext(ctx).Model(&model.Message{}).Where("status = ?", 2).Count(&replied).Error

	return map[string]int64{
		"unprocessed": unprocessed,
		"processed":   processed,
		"replied":     replied,
	}, nil
}


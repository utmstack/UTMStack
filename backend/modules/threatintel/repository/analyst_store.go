package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/utmstack/utmstack/backend/modules/threatintel/domain"
)

type AnalystStore struct {
	db *gorm.DB
}

// NewAnalystStore builds the GORM-backed conversation store.
func NewAnalystStore(db *gorm.DB) *AnalystStore {
	return &AnalystStore{db: db}
}

func (s *AnalystStore) Create(ctx context.Context, conv *domain.AnalystConversation) error {
	now := time.Now().UTC()
	conv.CreatedDate = now
	conv.ModifiedDate = now
	return s.db.WithContext(ctx).Create(conv).Error
}

func (s *AnalystStore) List(ctx context.Context, tenantID, userID string) ([]domain.AnalystConversation, error) {
	var rows []domain.AnalystConversation
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND user_id = ?", tenantID, userID).
		Order("updated_at DESC").
		Find(&rows).Error
	return rows, err
}

func (s *AnalystStore) Get(ctx context.Context, tenantID, userID, id string) (*domain.AnalystConversation, error) {
	var row domain.AnalystConversation
	err := s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).
		First(&row).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *AnalystStore) Save(ctx context.Context, conv *domain.AnalystConversation) error {
	now := time.Now().UTC()
	conv.UpdatedAt = now
	conv.ModifiedDate = now
	return s.db.WithContext(ctx).
		Model(&domain.AnalystConversation{}).
		Where("id = ? AND tenant_id = ? AND user_id = ?", conv.ID, conv.TenantID, conv.UserID).
		Updates(map[string]any{
			"title":         conv.Title,
			"turns":         conv.Turns,
			"history":       conv.History,
			"updated_at":    conv.UpdatedAt,
			"modified_date": conv.ModifiedDate,
		}).Error
}

func (s *AnalystStore) Delete(ctx context.Context, tenantID, userID, id string) error {
	return s.db.WithContext(ctx).
		Where("id = ? AND tenant_id = ? AND user_id = ?", id, tenantID, userID).
		Delete(&domain.AnalystConversation{}).Error
}

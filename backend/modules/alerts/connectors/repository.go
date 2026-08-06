package connectors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AlertRepository interface {
	UpdateStatus(ctx context.Context, alertIDs []string, status int, statusLabel, observation string) error
	UpdateStatusAndTag(ctx context.Context, alertIDs []string) error // adds "False positive" tag
	UpdateNotes(ctx context.Context, alertID, notes string) error
	UpdateAssignee(ctx context.Context, alertID, assignee string) error
	UpdateTags(ctx context.Context, alertIDs []string, tags []string) error
	ConvertToIncident(ctx context.Context, alertIDs []string, name string, id int, createdAt time.Time, createdBy, source string) error
	CountOpenAlerts(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status int) (int64, error)
	SearchByIDs(ctx context.Context, alertIDs []string) ([]domain.UtmAlert, error)
	ListEchoes(ctx context.Context, parentID string, from, size int, sortBy, sortOrder string) ([]domain.UtmAlert, int64, error)
	GetRawByID(ctx context.Context, alertID string) (json.RawMessage, error)
	RelatedLogRefs(ctx context.Context, steps []domain.CorrelationStep, anchorTS time.Time, maxSize int) (refs []domain.LogRef, truncated bool, err error)
}

type HistoryEntry struct {
	AlertID  string
	User     string
	Message  string
	Action   domain.AlertAction
	NewValue string // JSON-serialised changed fields
	At       time.Time
}

type HistoryRecorder interface {
	Record(ctx context.Context, entries []HistoryEntry) error
}

type AlertTagRepository interface {
	Create(ctx context.Context, tag domain.UtmAlertTag) (*domain.UtmAlertTag, error)
	Update(ctx context.Context, tag domain.UtmAlertTag) (*domain.UtmAlertTag, error)
	List(ctx context.Context, filters dto.AlertTagFilters) ([]domain.UtmAlertTag, int64, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmAlertTag, error)
	FindByIDs(ctx context.Context, ids []int64) ([]domain.UtmAlertTag, error)
	Delete(ctx context.Context, id uint64) error
}

type AlertTagRuleRepository interface {
	Create(ctx context.Context, rule domain.UtmAlertTagRule) (*domain.UtmAlertTagRule, error)
	Update(ctx context.Context, rule domain.UtmAlertTagRule) (*domain.UtmAlertTagRule, error)
	List(ctx context.Context, filters dto.AlertTagRuleFilters) ([]domain.UtmAlertTagRule, int64, error)
	GetByID(ctx context.Context, id uint64) (*domain.UtmAlertTagRule, error)
	FindByIDs(ctx context.Context, ids []int64) ([]domain.UtmAlertTagRule, error)
	Delete(ctx context.Context, id uint64) error
	FindAllActive(ctx context.Context) ([]domain.UtmAlertTagRule, error)
}

type AdversaryRepository interface {
	AdversaryAggs(ctx context.Context, filters []common_models.FilterType) (json.RawMessage, error)
}

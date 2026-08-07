package connectors

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type AlertRepository interface {
	UpdateStatus(ctx context.Context, alertIDs []string, status domain.AlertStatus, observation string, addFalsePositiveTag bool, history []HistoryEntry) error
	UpdateNotes(ctx context.Context, alertID, notes string, history []HistoryEntry) error
	UpdateAssignee(ctx context.Context, alertID, assignee string, history []HistoryEntry) error
	UpdateTags(ctx context.Context, alertIDs []string, tags []string, history []HistoryEntry) error
	ConvertToIncident(ctx context.Context, alertIDs []string, name, id string, createdAt time.Time, createdBy, source string, history []HistoryEntry) error
	CountOpenAlerts(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status domain.AlertStatus) (int64, error)
	SearchByIDs(ctx context.Context, alertIDs []string) ([]domain.UtmAlert, error)
	ListEchoes(ctx context.Context, parentID string, from, size int, sortBy, sortOrder string) ([]domain.UtmAlert, int64, error)
	GetRawByID(ctx context.Context, alertID string) (json.RawMessage, error)
	RelatedLogRefs(ctx context.Context, steps []domain.CorrelationStep, anchorTS time.Time, maxSize int) (refs []domain.LogRef, truncated bool, err error)
}

type HistoryEntry struct {
	AlertID  string
	User     string
	Action   domain.AlertAction
	NewValue string // JSON-serialised changed fields
	At       time.Time
}

type AlertTagRepository interface {
	Create(ctx context.Context, tag domain.AlertTag) (*domain.AlertTag, error)
	Update(ctx context.Context, tag domain.AlertTag) (*domain.AlertTag, error)
	List(ctx context.Context, filters dto.AlertTagFilters) ([]domain.AlertTag, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AlertTag, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.AlertTag, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type AlertTagRuleRepository interface {
	Create(ctx context.Context, rule domain.AlertTagRule) (*domain.AlertTagRule, error)
	Update(ctx context.Context, rule domain.AlertTagRule) (*domain.AlertTagRule, error)
	List(ctx context.Context, filters dto.AlertTagRuleFilters) ([]domain.AlertTagRule, int64, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AlertTagRule, error)
	FindByIDs(ctx context.Context, ids []uuid.UUID) ([]domain.AlertTagRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	FindAllActive(ctx context.Context) ([]domain.AlertTagRule, error)
}

type AdversaryGroup struct {
	Key    string
	Alerts []json.RawMessage
}

type AdversaryRepository interface {
	AdversaryGroups(ctx context.Context, filters []common_models.FilterType) ([]AdversaryGroup, error)
}

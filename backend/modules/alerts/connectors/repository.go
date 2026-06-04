package connectors

import (
	"context"
	"time"

	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

type AlertRepository interface {
	UpdateStatus(ctx context.Context, alertIDs []string, status int, statusLabel, observation string) error
	UpdateStatusAndTag(ctx context.Context, alertIDs []string) error // adds "False positive" tag
	UpdateNotes(ctx context.Context, alertID, notes string) error
	UpdateTags(ctx context.Context, alertIDs []string, tags []string) error
	ConvertToIncident(ctx context.Context, alertIDs []string, name string, id int, createdAt time.Time, createdBy, source string) error
	CountOpenAlerts(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status int) (int64, error)
	SearchByIDs(ctx context.Context, alertIDs []string) ([]domain.UtmAlert, error)
	ApplyTagRule(ctx context.Context, rule domain.UtmAlertTagRule, tagsForInsert []string, evaluationStart time.Time) error
	ReleaseToOpen(ctx context.Context, evaluationStart time.Time) error
	AssignAssetGroups(ctx context.Context, mapping map[string]AssetGroupRef) error
	SearchByStatusAndTimestamp(ctx context.Context, status int, evaluationStart time.Time, size int) ([]domain.UtmAlert, error)
	SearchByRuleConditionsAndTimestamp(ctx context.Context, rule domain.UtmAlertTagRule, evaluationStart time.Time, size int) ([]domain.UtmAlert, error)
}

type AssetGroupRef struct {
	ID   int64
	Name string
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

package connectors

import (
	"context"

	alertdomain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type ScoringUsecase interface {
	ScoreAlert(ctx context.Context, alertID string) (*domain.Score, error)
}

type AlertSearch interface {
	FetchByID(ctx context.Context, id string) (*alertdomain.UtmAlert, error)
	Count(ctx context.Context, filters []common_models.FilterType) (int64, error)
	Recent(ctx context.Context, filters []common_models.FilterType, n int, oldestFirst bool) ([]alertdomain.UtmAlert, error)
}

type AssetInfo struct {
	Hostname        string
	OS              string
	OSVersion       string
	IPs             []string
	Status          string // ONLINE | OFFLINE | UNKNOWN | ""
	Confidentiality int    // asset CIA sensitivity (0–3), from the datasource
	Integrity       int
	Availability    int
	HasSensitivity  bool // true when CIA was resolved from a registered datasource
}

type AssetLookup interface {
	Lookup(ctx context.Context, hostname string) AssetInfo
}

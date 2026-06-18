package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
	osdto "github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

type ScoringUsecase interface {
	ScoreAlert(ctx context.Context, alertID string) (*domain.Score, error)
}

type AlertSearch interface {
	Search(ctx context.Context, filters []common_models.FilterType, top int, indexPattern string, includeChildren bool, page, size int, sortBy, sortOrder string) ([]map[string]any, int64, error)
	PropertyValuesWithCount(ctx context.Context, req osdto.PropertyValuesWithCountRequest) (map[string]int64, error)
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

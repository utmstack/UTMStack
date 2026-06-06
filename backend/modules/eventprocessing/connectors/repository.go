package connectors

import (
	"context"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

type IngestionStatsRepository interface {
	TotalsByField(ctx context.Context, field, statusType, from, to string, top int) ([]dto.IngestionStatsBucket, int64, error)
	Timeline(ctx context.Context, statusType, interval, from, to string) ([]dto.TimelinePoint, error)
	TimelineByField(ctx context.Context, field, statusType, interval, from, to string, top int) ([]dto.TimelineSeries, error)
}

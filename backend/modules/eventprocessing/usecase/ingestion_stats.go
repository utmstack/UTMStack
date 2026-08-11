package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

var ErrInvalidStatsParam = errors.New("invalid ingestion-stats parameter")

const (
	defaultStatsWindow = "24h"
	defaultTop         = 100
	maxTop             = 1000
)

var statusToType = map[string]string{
	"received":            "enqueue_success",
	"parsing_dropped":     "parsing_dropped",
	"analysis_dropped":    "analysis_dropped",
	"correlation_dropped": "correlation_dropped",
	"all":                 "",
}

type ingestionStatsUsecase struct {
	repo connectors.IngestionStatsRepository
}

func NewIngestionStatsUsecase(repo connectors.IngestionStatsRepository) connectors.IngestionStatsUsecase {
	return &ingestionStatsUsecase{repo: repo}
}

func (u *ingestionStatsUsecase) Totals(ctx context.Context, groupBy, status, from, to string, top int) (*dto.IngestionStatsResponse, error) {
	field, err := resolveGroupBy(groupBy, true)
	if err != nil {
		return nil, err
	}
	status, statusType, err := resolveStatus(status)
	if err != nil {
		return nil, err
	}
	from, to = resolveWindow(from, to)
	top = clampTop(top)

	buckets, total, err := u.repo.TotalsByField(ctx, field, statsFilters(statusType, from, to, ""), top)
	if err != nil {
		return nil, err
	}
	return &dto.IngestionStatsResponse{
		GroupBy: groupBy,
		Status:  status,
		From:    from,
		To:      to,
		Total:   total,
		Buckets: buckets,
	}, nil
}

func (u *ingestionStatsUsecase) Timeline(ctx context.Context, groupBy, status, interval, from, to string, top int, dataSource string) (*dto.IngestionTimelineResponse, error) {
	// groupBy is optional for the timeline (empty → a single flat series).
	field, err := resolveGroupBy(groupBy, false)
	if err != nil {
		return nil, err
	}
	status, statusType, err := resolveStatus(status)
	if err != nil {
		return nil, err
	}
	from, to = resolveWindow(from, to)
	interval = resolveInterval(interval, from, to)
	top = clampTop(top)

	resp := &dto.IngestionTimelineResponse{
		Status:   status,
		GroupBy:  groupBy,
		Interval: interval,
		From:     from,
		To:       to,
	}

	if field == "" {
		points, err := u.repo.Timeline(ctx, statsFilters(statusType, from, to, dataSource), interval)
		if err != nil {
			return nil, err
		}
		resp.Points = points
		return resp, nil
	}

	series, err := u.repo.TimelineByField(ctx, field, statsFilters(statusType, from, to, dataSource), interval, top)
	if err != nil {
		return nil, err
	}
	resp.Series = series
	return resp, nil
}

func resolveGroupBy(groupBy string, required bool) (string, error) {
	switch groupBy {
	case "dataSource", "dataType":
		return groupBy, nil
	case "":
		if required {
			return "", ErrInvalidStatsParam
		}
		return "", nil
	default:
		return "", ErrInvalidStatsParam
	}
}

func resolveStatus(status string) (string, string, error) {
	if status == "" {
		status = "received"
	}
	t, ok := statusToType[status]
	if !ok {
		return "", "", ErrInvalidStatsParam
	}
	return status, t, nil
}

func resolveWindow(from, to string) (string, string) {
	if to == "" {
		to = "now"
	}
	if from == "" {
		from = "now-" + defaultStatsWindow
	}
	return from, to
}

func resolveInterval(interval, from, to string) string {
	if interval != "" && interval != "auto" {
		return interval
	}
	fromT, err1 := time.Parse(time.RFC3339, from)
	toT, err2 := time.Parse(time.RFC3339, to)
	if err1 != nil || err2 != nil {
		return "1h"
	}
	switch d := toT.Sub(fromT); {
	case d <= 2*time.Hour:
		return "5m"
	case d <= 48*time.Hour:
		return "1h"
	case d <= 14*24*time.Hour:
		return "1d"
	default:
		return "7d"
	}
}

func clampTop(top int) int {
	if top <= 0 {
		return defaultTop
	}
	if top > maxTop {
		return maxTop
	}
	return top
}
func statsFilters(statusType, from, to, dataSource string) []common_models.FilterType {
	filters := []common_models.FilterType{{
		Field:    "@timestamp",
		Operator: common_models.OpIsBetween,
		Value:    []any{from, to},
	}}
	if statusType != "" {
		filters = append(filters, common_models.FilterType{
			Field: "type", Operator: common_models.OpEquals, Value: statusType,
		})
	}
	if dataSource != "" {
		filters = append(filters, common_models.FilterType{
			Field: "dataSource", Operator: common_models.OpEquals, Value: dataSource,
		})
	}
	return filters
}

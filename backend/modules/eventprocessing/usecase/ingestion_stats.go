package usecase

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/utmstack/backend/modules/eventprocessing/connectors"
	"github.com/utmstack/utmstack/backend/modules/eventprocessing/dto"
)

var ErrInvalidStatsParam = errors.New("invalid ingestion-stats parameter")

const (
	defaultStatsWindow = 24 * time.Hour
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
	fromT, toT := resolveWindow(from, to)
	top = clampTop(top)

	buckets, totals, err := u.repo.TotalsByField(ctx, field, connectors.IngestionStatsQuery{
		From: fromT, To: toT, Type: statusType,
	}, top)
	if err != nil {
		return nil, err
	}
	return &dto.IngestionStatsResponse{
		GroupBy:    groupBy,
		Status:     status,
		From:       fromT.Format(time.RFC3339),
		To:         toT.Format(time.RFC3339),
		Total:      totals.Events,
		TotalBytes: totals.Bytes,
		Buckets:    buckets,
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
	fromT, toT := resolveWindow(from, to)
	interval = resolveInterval(interval, fromT, toT)
	top = clampTop(top)

	resp := &dto.IngestionTimelineResponse{
		Status:   status,
		GroupBy:  groupBy,
		Interval: interval,
		From:     fromT.Format(time.RFC3339),
		To:       toT.Format(time.RFC3339),
	}

	q := connectors.IngestionStatsQuery{
		From: fromT, To: toT, Type: statusType, DataSource: dataSource,
	}

	if field == "" {
		points, err := u.repo.Timeline(ctx, q, interval)
		if err != nil {
			return nil, err
		}
		resp.Points = points
		return resp, nil
	}

	series, err := u.repo.TimelineByField(ctx, field, q, interval, top)
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

// resolveWindow turns what the request asked for into real bounds. Both ends
// accept RFC3339 or the relative form the defaults use ("now", "now-24h"): a
// window nobody resolved is a window nobody applied, and the panel would then
// answer about all of history while its label said 24 hours.
func resolveWindow(from, to string) (time.Time, time.Time) {
	now := time.Now().UTC()

	toT, ok := parseWhen(to, now)
	if !ok {
		toT = now
	}
	fromT, ok := parseWhen(from, now)
	if !ok {
		fromT = toT.Add(-defaultStatsWindow)
	}
	if fromT.After(toT) {
		fromT, toT = toT, fromT
	}
	return fromT, toT
}

// parseWhen understands an absolute instant and the "now-<duration>" form.
func parseWhen(s string, now time.Time) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	if s == "now" {
		return now, true
	}
	if rest, found := strings.CutPrefix(s, "now-"); found {
		if d, err := parseWindow(rest); err == nil {
			return now.Add(-d), true
		}
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// parseWindow extends time.ParseDuration with the units a window is written in.
func parseWindow(s string) (time.Duration, error) {
	if len(s) > 1 {
		switch s[len(s)-1] {
		case 'd', 'D':
			n, err := strconv.Atoi(s[:len(s)-1])
			if err == nil && n > 0 {
				return time.Duration(n) * 24 * time.Hour, nil
			}
		case 'w', 'W':
			n, err := strconv.Atoi(s[:len(s)-1])
			if err == nil && n > 0 {
				return time.Duration(n) * 7 * 24 * time.Hour, nil
			}
		}
	}
	return time.ParseDuration(s)
}

func resolveInterval(interval string, from, to time.Time) string {
	if interval != "" && interval != "auto" {
		return interval
	}
	switch d := to.Sub(from); {
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

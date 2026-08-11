package usecase

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alerts/dto"
)

const relatedMaxSize = 1000

func (u *alertUsecase) RelatedLogs(ctx context.Context, alertID string) (*dto.RelatedLogsResponse, error) {
	raw, err := u.repo.GetRawByID(ctx, alertID)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, domain.ErrAlertNotFound
	}

	ruleName := gjson.GetBytes(raw, "name").String()
	tsStr := gjson.GetBytes(raw, "@timestamp").String()
	lastEvent := gjson.GetBytes(raw, "lastEvent")
	anchorTS := parseAlertTime(tsStr)

	var ids []string
	seen := make(map[string]struct{})
	add := func(id string) {
		if id == "" {
			return
		}
		if _, dup := seen[id]; dup {
			return
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	earliest, latest := anchorTS, anchorTS
	span := func(t time.Time) {
		if t.IsZero() {
			return
		}
		if t.Before(earliest) {
			earliest = t
		}
		if t.After(latest) {
			latest = t
		}
	}
	for _, ev := range gjson.GetBytes(raw, "events").Array() {
		add(ev.Get("id").String())
		span(eventTime(ev))
	}
	add(lastEvent.Get("id").String())
	span(eventTime(lastEvent))

	resp := &dto.RelatedLogsResponse{
		DataType: gjson.GetBytes(raw, "lastEvent.dataType").String(),
		TimeTo:   latest.Add(5 * time.Minute).UTC().Format(time.RFC3339),
		TimeFrom: earliest.Add(-time.Hour).UTC().Format(time.RFC3339),
	}

	var afterRaw json.RawMessage
	var found bool
	if u.resolver != nil {
		afterRaw, found = u.resolver.AfterEventsByRuleName(ruleName)
	}
	resp.RuleMatched = found

	if !found || len(afterRaw) == 0 || string(afterRaw) == "null" {
		resp.IDs = ids
		resp.Total = len(ids)
		return resp, nil
	}

	steps, err := resolveSteps(afterRaw, lastEvent)
	if err != nil {
		resp.IDs = ids
		resp.Total = len(ids)
		return resp, nil
	}

	refs, truncated, err := u.repo.RelatedLogRefs(ctx, steps, anchorTS, relatedMaxSize)
	if err != nil {
		return nil, err
	}
	for _, ref := range refs {
		add(ref.ID)
	}

	if w := maxWithin(steps); w > time.Hour {
		resp.TimeFrom = earliest.Add(-w).UTC().Format(time.RFC3339)
	}

	resp.IDs = ids
	resp.Total = len(ids)
	resp.Truncated = truncated
	return resp, nil
}

type rawExpr struct {
	Field    string          `json:"field"`
	Operator string          `json:"operator"`
	Value    json.RawMessage `json:"value"`
}

type rawStep struct {
	IndexPattern string    `json:"indexPattern"`
	With         []rawExpr `json:"with"`
	Or           []rawStep `json:"or"`
	Within       string    `json:"within"`
	Count        uint64    `json:"count"`
}

func resolveSteps(raw json.RawMessage, lastEvent gjson.Result) ([]domain.CorrelationStep, error) {
	var rs []rawStep
	if err := json.Unmarshal(raw, &rs); err != nil {
		return nil, err
	}
	return mapRawSteps(rs, lastEvent), nil
}

func mapRawSteps(rs []rawStep, le gjson.Result) []domain.CorrelationStep {
	out := make([]domain.CorrelationStep, 0, len(rs))
	for _, s := range rs {
		step := domain.CorrelationStep{
			IndexPattern: s.IndexPattern,
			Within:       s.Within,
			Count:        s.Count,
			Or:           mapRawSteps(s.Or, le),
		}
		for _, e := range s.With {
			step.With = append(step.With, domain.CorrelationExpr{
				Field:    e.Field,
				Operator: e.Operator,
				Value:    resolveValue(e.Value, le),
			})
		}
		out = append(out, step)
	}
	return out
}

func resolveValue(raw json.RawMessage, le gjson.Result) any {
	if len(raw) == 0 {
		return nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		if strings.HasPrefix(s, "{{.") && strings.HasSuffix(s, "}}") {
			path := strings.TrimSuffix(strings.TrimPrefix(s, "{{."), "}}")
			v := le.Get(path)
			if !v.Exists() {
				return nil
			}
			return v.Value()
		}
		return s
	}
	var v any
	if err := json.Unmarshal(raw, &v); err == nil {
		return v
	}
	return nil
}

func maxWithin(steps []domain.CorrelationStep) time.Duration {
	var max time.Duration
	for i := range steps {
		if d, err := time.ParseDuration(steps[i].Within); err == nil && d > max {
			max = d
		}
		if w := maxWithin(steps[i].Or); w > max {
			max = w
		}
	}
	return max
}

func parseAlertTime(s string) time.Time {
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.UTC()
		}
	}
	return time.Now().UTC()
}

func eventTime(ev gjson.Result) time.Time {
	for _, f := range []string{"@timestamp", "timestamp", "deviceTime"} {
		s := ev.Get(f).String()
		if s == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
			if t, err := time.Parse(layout, s); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Time{}
}

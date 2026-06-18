package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	alertdomain "github.com/utmstack/utmstack/backend/modules/alerts/domain"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/connectors"
	"github.com/utmstack/utmstack/backend/modules/alertscoring/domain"
	osdto "github.com/utmstack/utmstack/backend/modules/opensearch/dto"
	"github.com/utmstack/utmstack/backend/pkg/common_models"
)

const (
	alertIndex   = "v11-alert-*"
	baselineDays = 30
	chainDays    = 7
	maxRelated   = 50

	weightP1 = 0.30
	weightP2 = 0.25
	weightP3 = 0.20
	weightP4 = 0.25

	thresholdComplete = 50 // ≤ → COMPLETE
	thresholdIncident = 75 // ≤ → IN_REVIEW, else INCIDENT
)

type scorer struct {
	search connectors.AlertSearch
	assets connectors.AssetLookup
}

func NewScorer(search connectors.AlertSearch, assets connectors.AssetLookup) connectors.ScoringUsecase {
	return &scorer{search: search, assets: assets}
}

func (s *scorer) ScoreAlert(ctx context.Context, alertID string) (*domain.Score, error) {
	if strings.TrimSpace(alertID) == "" {
		return nil, fmt.Errorf("alert_id is required")
	}
	alert, err := s.fetchAlertByID(ctx, alertID)
	if err != nil {
		return nil, err
	}
	if alert == nil {
		return nil, fmt.Errorf("alert %q not found", alertID)
	}

	anchor := alertAnchor(alert.Timestamp)
	baseStart, baseEnd := anchor.AddDate(0, 0, -baselineDays), anchor
	chainStart, chainEnd := anchor.AddDate(0, 0, -chainDays), anchor

	host := firstNonEmpty(sideHost(alert.Adversary), alert.DataSource)

	// Phase 2 gathering.
	nameCount := s.alertNameCount(ctx, alert.Name, baseStart, baseEnd)
	sameAsset := s.sameAssetCount(ctx, alert.Name, host, baseStart, baseEnd)
	pastAlerts := s.pastAlerts(ctx, alert.Name, baseStart, baseEnd)

	// Phase 4 gathering.
	related := s.relatedAlerts(ctx, alert.Adversary, chainStart, chainEnd)

	// Phase 3 asset enrichment (best-effort).
	assetHost := firstNonEmpty(alert.DataSource, host)
	asset := connectors.AssetInfo{}
	if s.assets != nil {
		asset = s.assets.Lookup(ctx, assetHost)
	}
	if asset.Hostname == "" {
		asset.Hostname = assetHost
	}

	p1 := phase1Intrinsic(alert)
	p2 := phase2Baseline(nameCount, sameAsset, pastAlerts)
	p3 := phase3Asset(alert.DataType, asset)
	p4 := phase4Chain(alert.ID, alert.Name, related)

	rawFinal := float64(p1.Score)*weightP1 + float64(p2.Score)*weightP2 +
		float64(p3.Score)*weightP3 + float64(p4.Score)*weightP4
	final := clampScore(rawFinal)
	decision := decide(final)

	return &domain.Score{
		AlertID:    alert.ID,
		AlertName:  alert.Name,
		FinalScore: final,
		Decision:   decision,
		Weights: map[string]float64{
			"intrinsic": weightP1, "baseline": weightP2, "asset": weightP3, "chain": weightP4,
		},
		Phases: map[string]domain.PhaseResult{
			"intrinsic": p1, "baseline": p2, "asset": p3, "chain": p4,
		},
		Summary: buildSummary(alert, final, decision, p2, p3, p4),
	}, nil
}

// ---------------------------------------------------------------------------
// OpenSearch gathering
// ---------------------------------------------------------------------------

func (s *scorer) fetchAlertByID(ctx context.Context, id string) (*alertdomain.UtmAlert, error) {
	filters := []common_models.FilterType{{Field: "id", Operator: common_models.OpEquals, Value: id}}
	hits, _, err := s.search.Search(ctx, filters, 1, alertIndex, false, 1, 1, "@timestamp", "desc")
	if err != nil {
		return nil, fmt.Errorf("fetch alert: %w", err)
	}
	if len(hits) == 0 {
		return nil, nil
	}
	a := hitToAlert(hits[0])
	return &a, nil
}

func (s *scorer) alertNameCount(ctx context.Context, name string, start, end time.Time) int64 {
	if name == "" {
		return 0
	}
	counts, err := s.search.PropertyValuesWithCount(ctx, propValuesReq(start, end))
	if err != nil {
		return 0
	}
	return counts[name]
}

func (s *scorer) sameAssetCount(ctx context.Context, name, host string, start, end time.Time) int64 {
	if name == "" || host == "" {
		return 0
	}
	filters := []common_models.FilterType{
		{Field: "name", Operator: common_models.OpEquals, Value: name},
		{Field: "adversary.host", Operator: common_models.OpEquals, Value: host},
		tsBetween(start, end),
	}
	_, total, err := s.search.Search(ctx, filters, 1000, alertIndex, false, 1, 1, "@timestamp", "desc")
	if err != nil {
		return 0
	}
	return total
}

func (s *scorer) pastAlerts(ctx context.Context, name string, start, end time.Time) []alertdomain.UtmAlert {
	if name == "" {
		return nil
	}
	filters := []common_models.FilterType{
		{Field: "name", Operator: common_models.OpEquals, Value: name},
		tsBetween(start, end),
	}
	hits, _, err := s.search.Search(ctx, filters, maxRelated, alertIndex, false, 1, maxRelated, "@timestamp", "desc")
	if err != nil {
		return nil
	}
	return hitsToAlerts(hits)
}

func (s *scorer) relatedAlerts(ctx context.Context, adv *alertdomain.Side, start, end time.Time) []alertdomain.UtmAlert {
	if adv == nil {
		return nil
	}
	pivots := []struct{ field, value string }{
		{"adversary.host", adv.Host},
		{"adversary.user", adv.User},
		{"adversary.ip", adv.IP},
	}
	seen := map[string]bool{}
	var out []alertdomain.UtmAlert
	for _, p := range pivots {
		if p.value == "" {
			continue
		}
		filters := []common_models.FilterType{
			{Field: p.field, Operator: common_models.OpEquals, Value: p.value},
			tsBetween(start, end),
		}
		hits, _, err := s.search.Search(ctx, filters, maxRelated, alertIndex, false, 1, maxRelated, "@timestamp", "asc")
		if err != nil {
			continue
		}
		for _, a := range hitsToAlerts(hits) {
			if a.ID != "" && seen[a.ID] {
				continue
			}
			seen[a.ID] = true
			out = append(out, a)
		}
	}
	return out
}

func propValuesReq(start, end time.Time) osdto.PropertyValuesWithCountRequest {
	return osdto.PropertyValuesWithCountRequest{
		Index:        alertIndex,
		Field:        "name.keyword",
		Top:          5000,
		OrderByCount: true,
		Filters:      []common_models.FilterType{tsBetween(start, end)},
	}
}

// ---------------------------------------------------------------------------
// aggregation + summary
// ---------------------------------------------------------------------------

func decide(final int) string {
	switch {
	case final <= thresholdComplete:
		return domain.DecisionComplete
	case final <= thresholdIncident:
		return domain.DecisionInReview
	default:
		return domain.DecisionIncident
	}
}

func buildSummary(a *alertdomain.UtmAlert, final int, decision string, p2, p3, p4 domain.PhaseResult) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Score %d/100 (%s).", final, decision)

	if stages, ok := p4.Breakdown["killChainStages"].([]string); ok && len(stages) >= 2 {
		fmt.Fprintf(&b, " %d-stage kill chain (%s).", len(stages), strings.Join(stages, " → "))
	} else if cnt, ok := p4.Breakdown["relatedAlertCount"].(int); ok && cnt == 0 {
		b.WriteString(" Isolated event.")
	}

	host := firstNonEmpty(a.DataSource, sideHost(a.Adversary))
	if host != "" {
		fmt.Fprintf(&b, " Asset %s", host)
		if lbl, ok := p3.Breakdown["agentLabel"].(string); ok && lbl == "Offline" {
			b.WriteString(" (agent offline)")
		}
		b.WriteString(".")
	}

	if detail, ok := p2.Breakdown["resolutionDetail"].(string); ok {
		switch detail {
		case "mostly_false_positives", "significant_false_positives":
			b.WriteString(" This alert type was historically a false positive.")
		case "mostly_incidents", "some_incidents":
			b.WriteString(" This alert type historically escalated to incidents.")
		}
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func tsBetween(start, end time.Time) common_models.FilterType {
	return common_models.FilterType{
		Field:    "@timestamp",
		Operator: common_models.OpIsBetween,
		Value:    []string{start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339)},
	}
}

func hitsToAlerts(hits []map[string]any) []alertdomain.UtmAlert {
	out := make([]alertdomain.UtmAlert, 0, len(hits))
	for _, h := range hits {
		out = append(out, hitToAlert(h))
	}
	return out
}

func hitToAlert(hit map[string]any) alertdomain.UtmAlert {
	var a alertdomain.UtmAlert
	raw, err := json.Marshal(hit)
	if err != nil {
		return a
	}
	_ = json.Unmarshal(raw, &a)
	return a
}

func sideHost(s *alertdomain.Side) string {
	if s == nil {
		return ""
	}
	return s.Host
}

func alertAnchor(ts string) time.Time {
	if t, ok := parseAlertTime(ts); ok {
		return t
	}
	return time.Now().UTC()
}

func parseAlertTime(ts string) (time.Time, bool) {
	ts = strings.TrimSpace(ts)
	if ts == "" {
		return time.Time{}, false
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02T15:04:05.999Z07:00", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, ts); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func clampScore(raw float64) int {
	r := int(math.Round(raw))
	if r < 0 {
		return 0
	}
	if r > 100 {
		return 100
	}
	return r
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round1(v float64) float64 { return math.Round(v*10) / 10 }

func itoa(v int) string { return strconv.Itoa(v) }

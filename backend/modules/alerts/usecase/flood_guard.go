package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/threatwinds/go-sdk/catcher"

	"github.com/utmstack/utmstack/backend/modules/alerts/connectors"
)

const floodNotificationTemplate = "Correlation rule '%s' generated %d open, un-deduplicated alerts from data source '%s' in the last %dh and was automatically disabled to prevent alert flooding. If this volume is expected, use an Alert Tag Rule to mark it 'False positive', or add deduplicateBy/groupBy to the rule, then re-enable it."

// FloodGuard is the safety net against alert fatigue: when a correlation rule floods one tenant's
// alerts with individual, non-deduplicated alerts from the same data source, it disables the rule for
// that tenant and tells them through the notification bell.
//
// Every replica of the backend runs it. That is safe because Deactivate says whether this call is the
// one that changed the rule, and only that one notifies.
type FloodGuard struct {
	alerts   connectors.FloodRepository
	rules    connectors.FloodRuleSwitch
	notifier connectors.FloodNotifier
	config   func() FloodConfig
}

// NewFloodGuard builds the guard. config is read on every pass.
func NewFloodGuard(alerts connectors.FloodRepository, rules connectors.FloodRuleSwitch, notifier connectors.FloodNotifier, config func() FloodConfig) *FloodGuard {
	return &FloodGuard{alerts: alerts, rules: rules, notifier: notifier, config: config}
}

// Start checks once and then every configured interval until ctx is done.
func (g *FloodGuard) Start(ctx context.Context) {
	if g == nil {
		return
	}

	timer := time.NewTimer(g.config().Interval)
	defer timer.Stop()

	g.evaluateOnce(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			g.evaluateOnce(ctx)
			timer.Reset(g.config().Interval)
		}
	}
}

func (g *FloodGuard) evaluateOnce(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			_ = catcher.Error("flood guard: recovered from a panic", fmt.Errorf("%v", r), nil)
		}
	}()

	cfg := g.config()
	if !cfg.Enabled {
		return
	}

	buckets, err := g.alerts.FloodBuckets(ctx, cfg.window())
	if err != nil {
		_ = catcher.Error("flood guard: failed to count the alerts", err, nil)

		return
	}

	for _, b := range buckets {
		if b.Count <= cfg.Threshold {
			continue
		}

		if b.TenantID == "" {
			catcher.Warn("flood guard: dropping a flooding bucket with no tenant", map[string]any{
				"ruleName": b.RuleName, "dataSource": b.DataSource, "count": b.Count,
			})

			continue
		}

		changed, err := g.deactivate(ctx, b.TenantID, b.RuleName)
		if err != nil {
			_ = catcher.Error("flood guard: failed to deactivate the rule", err, map[string]any{
				"tenantId": b.TenantID, "ruleName": b.RuleName, "dataSource": b.DataSource, "count": b.Count,
			})
		}

		if !changed {
			continue
		}

		message := fmt.Sprintf(floodNotificationTemplate, b.RuleName, b.Count, b.DataSource, cfg.WindowHours)
		if err := g.notifier.Warn(ctx, b.TenantID, message); err != nil {
			_ = catcher.Error("flood guard: failed to notify the tenant", err, map[string]any{
				"tenantId": b.TenantID, "ruleName": b.RuleName, "dataSource": b.DataSource,
			})
		}
	}
}

// deactivate turns off every active rule of the tenant whose name matches exactly, and reports whether
// any of them changed.
func (g *FloodGuard) deactivate(ctx context.Context, tenantID, ruleName string) (bool, error) {
	candidates, err := g.rules.FindByName(ctx, tenantID, ruleName)
	if err != nil {
		return false, err
	}

	var matches []connectors.FloodRule

	for _, r := range candidates {
		if strings.EqualFold(r.Name, ruleName) {
			matches = append(matches, r)
		}
	}

	if len(matches) == 0 {
		return false, nil
	}

	if len(matches) > 1 {
		catcher.Warn("flood guard: ambiguous rule name, disabling every exact match", map[string]any{
			"ruleName": ruleName, "matches": len(matches), "tenantId": tenantID,
		})
	}

	changed := false

	for _, r := range matches {
		if !r.Active {
			continue
		}

		did, err := g.rules.Deactivate(ctx, tenantID, r.RelPath)
		if err != nil {
			return changed, err
		}

		if did {
			changed = true
		}
	}

	return changed, nil
}

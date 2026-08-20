package main

import (
	"context"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

type searchFunc func(ctx context.Context, window time.Duration) ([]ruleBucket, error)

type disableNotifier interface {
	Deactivate(ctx context.Context, tenantID, ruleName string) (bool, error)
	Notify(ctx context.Context, tenantID, message string) error
}

type getConfig func() Config

func evaluateOnce(ctx context.Context, search searchFunc, client disableNotifier, getCfg getConfig) {
	cfg := getCfg()
	if !cfg.Enabled {
		return
	}

	buckets, err := search(ctx, cfg.window())
	if err != nil {
		_ = catcher.Error("rule-flood-guard: failed to search alert buckets", err, nil)
		return
	}

	for _, b := range buckets {
		if b.Count <= cfg.Threshold {
			continue
		}
		if b.TenantID == "" {
			catcher.Warn("rule-flood-guard: dropping a flooding bucket with no tenant", map[string]any{
				"ruleName": b.RuleName, "dataSource": b.DataSource, "count": b.Count,
			})
			continue
		}

		changed, err := client.Deactivate(ctx, b.TenantID, b.RuleName)
		if err != nil {
			_ = catcher.Error("rule-flood-guard: failed to deactivate rule", err, map[string]any{
				"tenantId": b.TenantID, "ruleName": b.RuleName,
				"dataSource": b.DataSource, "count": b.Count,
			})
		}
		if !changed {
			continue
		}

		msg := floodNotificationMessage(b.TenantID, b.RuleName, b.Count, b.DataSource, cfg.WindowHours)
		if err := client.Notify(ctx, b.TenantID, msg); err != nil {
			_ = catcher.Error("rule-flood-guard: failed to send notification", err, map[string]any{
				"tenantId": b.TenantID, "ruleName": b.RuleName, "dataSource": b.DataSource,
			})
		}
	}
}

func runLoop(ctx context.Context, search searchFunc, client disableNotifier, getCfg getConfig) {
	timer := time.NewTimer(getCfg().tickInterval())
	defer timer.Stop()

	evaluateOnce(ctx, search, client, getCfg)

	for {
		select {
		case <-timer.C:
			evaluateOnce(ctx, search, client, getCfg)
			timer.Reset(getCfg().tickInterval())
		case <-ctx.Done():
			return
		}
	}
}

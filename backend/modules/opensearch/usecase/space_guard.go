package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/threatwinds/go-sdk/catcher"
)

// diskOps is the subset of the gateway the space guard needs.
type diskOps interface {
	DiskUsedPercent(ctx context.Context) (float64, error)
	OldestIndices(ctx context.Context, pattern string) ([]string, error)
	DeleteIndex(ctx context.Context, indices []string) error
}

// removableChecker reports whether an index is in an ISM state that allows deletion.
type removableChecker interface {
	IsIndexRemovable(ctx context.Context, indexName string) bool
}

// SpaceNotifyFunc emits an operator notification. critical=true is the
// auto-delete (ERROR) level; false is the warning level.
type SpaceNotifyFunc func(ctx context.Context, critical bool, message string) error

// SpaceGuardOptions configures the disk-space guard.
type SpaceGuardOptions struct {
	WarnPercent   float64       // notify at/above this disk usage (e.g. 70)
	DeletePercent float64       // auto-delete oldest removable indices at/above this (e.g. 85)
	Interval      time.Duration // how often to check
	LogsPattern   string        // index pattern of log indices eligible for emergency deletion
}

const (
	spaceWarnCooldown   = 24 * time.Hour // throttle between warning notifications (in-memory)
	spaceDeletePause    = 10 * time.Second
	spaceRecoverMargin  = 5.0 // reclaim until DeletePercent - margin
	spaceMaxDeletesTick = 50  // safety cap on deletions per check
)

// SpaceGuard is a background watchdog that prevents OpenSearch from filling its
// disk and hitting the flood-stage watermark (which makes indices read-only and
// stops ingestion). It periodically checks per-node disk usage and, when high,
// notifies operators and — past the delete threshold — removes the oldest
// ISM-removable log indices until usage recovers.
//
// This is the safety net for ingestion spikes that outrun age-based ISM
// retention; steady-state cleanup is still handled by the ISM policy.
type SpaceGuard struct {
	ops       diskOps
	removable removableChecker
	notify    SpaceNotifyFunc
	opts      SpaceGuardOptions
	nextWarn  time.Time // in-memory warning throttle (no DB table needed)
}

func NewSpaceGuard(ops diskOps, removable removableChecker, notify SpaceNotifyFunc, opts SpaceGuardOptions) *SpaceGuard {
	if opts.Interval <= 0 {
		opts.Interval = time.Minute
	}
	return &SpaceGuard{ops: ops, removable: removable, notify: notify, opts: opts}
}

// Run blocks, checking disk usage every opts.Interval until ctx is cancelled.
func (g *SpaceGuard) Run(ctx context.Context) {
	catcher.Info(fmt.Sprintf("opensearch space guard started (warn=%.0f%% delete=%.0f%% every %s)",
		g.opts.WarnPercent, g.opts.DeletePercent, g.opts.Interval), nil)
	t := time.NewTicker(g.opts.Interval)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			g.check(ctx)
		}
	}
}

func (g *SpaceGuard) check(ctx context.Context) {
	pct, err := g.ops.DiskUsedPercent(ctx)
	if err != nil {
		_ = catcher.Error("space guard: disk usage check failed", err, nil)
		return
	}
	switch {
	case pct >= g.opts.DeletePercent:
		g.reclaim(ctx, pct)
	case pct >= g.opts.WarnPercent:
		g.warn(ctx, pct)
	}
}

func (g *SpaceGuard) warn(ctx context.Context, pct float64) {
	now := time.Now()
	if now.Before(g.nextWarn) {
		return // throttled
	}
	msg := fmt.Sprintf("OpenSearch disk usage is at %.0f%% (warning threshold %.0f%%). Review index retention before it reaches %.0f%% and indices are auto-deleted.",
		pct, g.opts.WarnPercent, g.opts.DeletePercent)
	if err := g.notify(ctx, false, msg); err != nil {
		_ = catcher.Error("space guard: notify failed", err, nil)
		return
	}
	g.nextWarn = now.Add(spaceWarnCooldown)
}

// reclaim deletes the oldest ISM-removable log indices until disk drops below
// the recovery target, to avoid hitting the flood-stage read-only watermark.
func (g *SpaceGuard) reclaim(ctx context.Context, startPct float64) {
	target := g.opts.DeletePercent - spaceRecoverMargin
	catcher.Warn(fmt.Sprintf("space guard: disk at %.0f%% (>= %.0f%%), reclaiming oldest indices toward %.0f%%",
		startPct, g.opts.DeletePercent, target), nil)

	indices, err := g.ops.OldestIndices(ctx, g.opts.LogsPattern)
	if err != nil {
		_ = catcher.Error("space guard: list oldest indices failed", err, nil)
		return
	}

	notified := false
	deleted := 0
	for _, idx := range indices {
		if deleted >= spaceMaxDeletesTick {
			break
		}
		select {
		case <-ctx.Done():
			return
		default:
		}

		pct, err := g.ops.DiskUsedPercent(ctx)
		if err != nil {
			_ = catcher.Error("space guard: disk recheck failed", err, nil)
			return
		}
		if pct < target {
			break // recovered
		}
		if !g.removable.IsIndexRemovable(ctx, idx) {
			continue // not in a deletable ISM state
		}

		if !notified {
			_ = g.notify(ctx, true, fmt.Sprintf("OpenSearch disk usage reached %.0f%%. Auto-deleting the oldest log indices to prevent ingestion from locking up.", startPct))
			notified = true
		}

		if err := g.ops.DeleteIndex(ctx, []string{idx}); err != nil {
			_ = catcher.Error("space guard: delete index failed", err, map[string]any{"index": idx})
			continue
		}
		deleted++
		catcher.Warn("space guard: deleted index to reclaim space", map[string]any{"index": idx})

		select {
		case <-ctx.Done():
			return
		case <-time.After(spaceDeletePause):
		}
	}

	if deleted == 0 {
		_ = catcher.Error("space guard: disk above delete threshold but no removable indices found — check ISM policy / retention", nil, nil)
	}
}

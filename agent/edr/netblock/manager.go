// edr/netblock/manager.go
package netblock

import (
	"context"
	"net/netip"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

type spooler interface{ Append(string) error }

type Deps struct {
	Cfg   config.EDRConfig
	Cache *cache.Cache
	Spool spooler
	Sys   SystemNets
}

type Health struct {
	Enabled, Enforce   bool
	Indicators         int
	ActiveFilters      int
	RecentBlocks       int
	AllowlistedDropped int
	FeedAgeSec         int64
	FeedStale          bool
}

type Manager struct {
	cfg   config.EDRConfig
	cache *cache.Cache
	spool spooler
	store *Store
	allow *Allowlist
	enf   *Enforcer
	feed  *Feed
	audit ConnAudit

	lastSync     atomic.Int64 // unix sec
	recentBlocks atomic.Int64
	dropped      atomic.Int64
	desired      atomic.Int64
	mu           sync.Mutex
}

func NewManager(d Deps) *Manager {
	blk, err := NewOSBlocker(config.ServiceName + "-Blocklist")
	if err != nil {
		logWarn("blocklist WFP init failed, enforcement disabled: %v", err)
		blk, _ = newDisabledBlocker()
	}
	m := &Manager{
		cfg:   d.Cfg,
		cache: d.Cache,
		spool: d.Spool,
		store: NewStore(),
		allow: BuildAllowlist(d.Cfg, d.Sys),
		enf:   NewEnforcer(blk, d.Cfg.Blocklist.Direction),
		audit: NewConnAudit(),
	}
	m.feed = NewFeed(d.Cfg, d.Cache, m.store, m.applyIndicators)
	return m
}

func (m *Manager) Run(ctx context.Context) {
	// Startup convergence: re-apply from cached indicators before the first fetch.
	if recs, err := m.cache.ListIndicators(); err == nil && len(recs) > 0 {
		inds := make([]Indicator, 0, len(recs))
		for _, r := range recs {
			if ind, err := ParseIndicator(r.Value, r.Type, r.Level); err == nil {
				inds = append(inds, ind)
			}
		}
		m.store.Swap(inds)
		m.applyIndicators(inds)
	}

	if m.cfg.Blocklist.Enforce {
		blockCh := make(chan BlockEvent, 256)
		go m.audit.Run(ctx, blockCh)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case be := <-blockCh:
					m.onBlockEvent(be)
				}
			}
		}()
	}

	m.feed.Run(ctx) // blocks until ctx.Done()
	_ = m.enf.Clear()
}

// applyIndicators recomputes the desired IP set (allowlist-filtered) and reconciles
// the enforcer. Domain/hostname indicators are ignored in Plan 1.
func (m *Manager) applyIndicators(inds []Indicator) {
	m.lastSync.Store(time.Now().Unix())
	var desired []netip.Addr
	var dropped int
	for _, ind := range inds {
		switch ind.Type {
		case "ip":
			if a, err := netip.ParseAddr(ind.Value); err == nil {
				if m.allow.Allowed(a) {
					dropped++
					continue
				}
				desired = append(desired, a)
			}
		case "cidr":
			// CIDR filters are applied directly by the blocker in Plan 2's
			// prefix path; Plan 1 enforces exact IPs. Expand is out of scope —
			// CIDRs are matched for *detection* via the store, blocked in Plan 2.
		}
	}
	m.dropped.Store(int64(dropped))
	if !m.cfg.Blocklist.Enforce {
		m.desired.Store(0)
		return
	}
	if err := m.enf.Apply(desired); err != nil {
		logWarn("blocklist enforce apply: %v", err)
	}
	m.desired.Store(int64(len(desired)))
}

// onBlockEvent turns a WFP drop into a branded event + audit record.
func (m *Manager) onBlockEvent(be BlockEvent) {
	m.recentBlocks.Add(1)
	ind, _ := m.store.Match(be.RemoteIP)
	var p *event.ProcInfo
	if be.PID != 0 || be.ProcessPath != "" {
		p = &event.ProcInfo{PID: be.PID, Image: be.ProcessPath}
	}
	ev := event.NewNetworkEvent(event.ActionBlocked, be.Direction, be.RemoteIP.String(),
		be.RemotePort, "", ind.Value, p)
	if js, err := ev.ToJSON(); err == nil {
		_ = m.spool.Append(js)
	}
	_ = m.cache.RecordNetBlock(cache.NetBlockRecord{
		Direction: be.Direction, RemoteIP: be.RemoteIP.String(), RemotePort: be.RemotePort,
		PID: be.PID, ProcessPath: be.ProcessPath, Matched: ind.Value, MatchedType: ind.Type,
		Level: ind.Level, Action: event.ActionBlocked,
	})
}

func (m *Manager) Health() Health {
	age := int64(0)
	stale := false
	if ls := m.lastSync.Load(); ls > 0 {
		age = time.Now().Unix() - ls
		stale = age > int64(24*3600) // >24h with no successful sync
	}
	return Health{
		Enabled:            m.cfg.Blocklist.Enabled,
		Enforce:            m.cfg.Blocklist.Enforce,
		Indicators:         m.store.Count(),
		ActiveFilters:      int(m.desired.Load()),
		RecentBlocks:       int(m.recentBlocks.Swap(0)),
		AllowlistedDropped: int(m.dropped.Load()),
		FeedAgeSec:         age,
		FeedStale:          stale,
	}
}

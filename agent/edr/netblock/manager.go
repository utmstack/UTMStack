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
	Enabled, Enforce    bool
	Indicators          int
	ActiveFilters       int
	RecentBlocks        int
	AllowlistedDropped  int
	FeedAgeSec          int64
	FeedStale           bool
	EnforcementDegraded bool // enforcement wanted but WFP unavailable (status must not lie)
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

	dnsFeed  DNSFeed
	reactive *Reactive

	lastSync     atomic.Int64 // unix sec
	recentBlocks atomic.Int64
	dropped      atomic.Int64
	desired      atomic.Int64
	// enforce mirrors cfg.Blocklist.Enforce for lock-free reads on the DNS-feed
	// goroutine (onDNSEvent), which must not touch m.cfg (guarded by m.mu, written
	// by Reload).
	enforce atomic.Bool
	wfpOK   bool // WFP init succeeded (real enforcement possible)
	// mu serializes the allowlist + enforcer across the feed's Run goroutine
	// (applyIndicators) and the service's reload goroutine (Reload).
	mu sync.Mutex
}

func NewManager(d Deps) *Manager {
	wfpOK := true
	blk, err := NewOSBlocker(config.ServiceName + "-Blocklist")
	if err != nil {
		logWarn("blocklist WFP init failed, enforcement disabled: %v", err)
		blk, _ = newDisabledBlocker()
		wfpOK = false
	}
	m := &Manager{
		cfg:   d.Cfg,
		cache: d.Cache,
		spool: d.Spool,
		store: NewStore(),
		allow: BuildAllowlist(d.Cfg, d.Sys),
		enf:   NewEnforcer(blk, d.Cfg.Blocklist.Direction),
		audit: NewConnAudit(),
		wfpOK: wfpOK,
	}
	m.feed = NewFeed(d.Cfg, d.Cache, m.store, m.applyIndicators)
	// DNS reactive path (Plan 2): resolve blacklisted domains → short-TTL WFP blocks
	// of the resolved IPs. Reactive wraps the SAME blocker the enforcer uses.
	m.dnsFeed = NewDNSFeed()
	ttl := time.Duration(d.Cfg.Blocklist.DomainBlockTTLMin) * time.Minute
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	m.reactive = NewReactive(blk, d.Cfg.Blocklist.Direction, ttl, time.Now)
	// Reactive and the feed enforcer share the SAME blocker. When a reactive block
	// expires but the feed has since adopted the addr, Sweep must hand off the shared
	// filter to the enforcer instead of deleting it (which the enforcer won't re-add).
	m.reactive.SetOnFeed(func(a netip.Addr) bool { _, ok := m.store.Match(a); return ok })
	m.enforce.Store(d.Cfg.Blocklist.Enforce)
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

	// Always start the observers; their EFFECTS are gated by m.enforce (read live
	// on each event), so an enforce live-reload takes hold without a restart. In
	// detect-only these are naturally inert: connaudit's onBlockEvent fires only on
	// real WFP drops (none when nothing is enforced), onDNSEvent emits an honest
	// action="detected" and blocks nothing, and reactive's sweep has nothing to expire.
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
	// DNS reactive blocking: watch resolutions, block IPs of blacklisted domains.
	go func() {
		_ = m.dnsFeed.Run(ctx, m.onDNSEvent)
	}()
	// Sweep expired reactive (domain-derived) blocks on a fixed cadence.
	go func() {
		t := time.NewTicker(time.Minute)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				m.reactive.Sweep()
			}
		}
	}()

	m.feed.Run(ctx) // blocks until ctx.Done()
	// Serialize the teardown clear with Reload/applyIndicators (both hold m.mu),
	// so a reload landing at shutdown can't race the enforcer's map.
	m.mu.Lock()
	_ = m.enf.Clear()
	m.mu.Unlock()
}

// applyIndicators recomputes the desired IP set (allowlist-filtered) and reconciles
// the enforcer. Domain/hostname indicators are ignored in Plan 1. It is the entry
// point for the feed's Run goroutine; m.mu serializes it against Reload (which runs
// on the service's reload goroutine).
func (m *Manager) applyIndicators(inds []Indicator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.applyLocked(inds)
}

// applyLocked reconciles the enforcer against the (allowlist-filtered) indicator
// set. The caller MUST hold m.mu — both applyIndicators and Reload funnel here so
// the mutex is taken exactly once per call (a non-reentrant sync.Mutex).
func (m *Manager) applyLocked(inds []Indicator) {
	m.lastSync.Store(time.Now().Unix())
	var desired []netip.Addr
	var desiredP []netip.Prefix
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
			if p, err := netip.ParsePrefix(ind.Value); err == nil {
				// Skip a blacklisted CIDR that overlaps the never-block set — not just
				// one whose network address is allowlisted. A CIDR that CONTAINS an
				// allowlisted exact IP (resolver, gateway, platform server) or overlaps
				// an allowlisted prefix would otherwise self-lockout the host.
				if m.allow.OverlapsAllowed(p.Masked()) {
					dropped++
					continue
				}
				desiredP = append(desiredP, p.Masked())
			}
		}
	}
	m.dropped.Store(int64(dropped))
	if !m.cfg.Blocklist.Enforce {
		// Detect-only / enforce-off must hold no live filters — clear any that a
		// prior enforce=true cycle applied before we stop reconciling. Clear resets
		// both the exact-IP and prefix maps (and Reset removes prefix filters).
		_ = m.enf.Clear()
		// Clear wiped ALL blocker filters, including reactive's shared ones — drop
		// reactive's stale expiry so a later re-enable re-blocks those IPs afresh.
		m.reactive.Reset()
		m.desired.Store(0)
		return
	}
	if err := m.enf.Apply(desired); err != nil {
		logWarn("blocklist enforce apply: %v", err)
	}
	if err := m.enf.ApplyPrefixes(desiredP); err != nil {
		logWarn("blocklist enforce apply prefixes: %v", err)
	}
	m.desired.Store(int64(len(desired)))
}

// Reload hot-applies a new config + system-nets set without a service restart: it
// rebuilds the never-block allowlist and re-derives the enforced IP set from the
// live indicator store. It runs on the service's reload goroutine while the feed
// drives applyIndicators on the manager's Run goroutine — m.mu serializes both
// around the allowlist + enforcer. Reading indicators from the cache happens before
// the lock (the cache has its own lock); the reconcile then runs under m.mu once.
func (m *Manager) Reload(cfg config.EDRConfig, sys SystemNets) {
	var inds []Indicator
	if recs, err := m.cache.ListIndicators(); err == nil {
		inds = make([]Indicator, 0, len(recs))
		for _, r := range recs {
			if ind, err := ParseIndicator(r.Value, r.Type, r.Level); err == nil {
				inds = append(inds, ind)
			}
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cfg = cfg
	m.enforce.Store(cfg.Blocklist.Enforce)
	m.allow = BuildAllowlist(cfg, sys)
	m.applyLocked(inds)
}

// onBlockEvent turns a WFP drop into a branded event + audit record. The WFP
// net-event feed delivers ALL system classify-drops, so we emit only for drops
// whose remote IP is on our blocklist — everything else (Windows Firewall drops,
// etc.) is ignored here.
func (m *Manager) onBlockEvent(be BlockEvent) {
	ind, ok := m.store.Match(be.RemoteIP)
	if !ok {
		return
	}
	m.recentBlocks.Add(1)
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

// onDNSEvent reactively blocks the IPs a blacklisted domain resolved to and emits
// a branded event. IPs already on the feed blocklist are left to the enforcer
// (not double-managed by Reactive). It runs on the DNS feed goroutine and touches
// only concurrency-safe surfaces (store atomics, reactive's own mutex, the enforce
// atomic, recentBlocks atomic, spool, cache) — never the enforcer or m.cfg — so it
// does NOT take m.mu.
func (m *Manager) onDNSEvent(ev DNSEvent) {
	ind, ok := m.store.MatchName(ev.Name)
	if !ok {
		return
	}
	m.recentBlocks.Add(1)
	var toBlock []netip.Addr
	for _, ip := range ev.IPs {
		if _, onList := m.store.Match(ip); !onList {
			toBlock = append(toBlock, ip)
		}
	}
	// Only actually block (and report "blocked") when enforcing; in detect-only we
	// emit an honest action="detected" domain event and apply no WFP filters.
	enf := m.enforce.Load()
	action := event.ActionDetected
	if enf && len(toBlock) > 0 {
		m.reactive.Block(toBlock)
		action = event.ActionBlocked
	}
	var p *event.ProcInfo
	if ev.PID != 0 {
		p = &event.ProcInfo{PID: ev.PID, Image: ev.Process}
	}
	firstIP := ""
	if len(ev.IPs) > 0 {
		firstIP = ev.IPs[0].String()
	}
	e := event.NewNetworkEvent(action, "outbound", firstIP, 0, ev.Name, ind.Value, p)
	if js, err := e.ToJSON(); err == nil {
		_ = m.spool.Append(js)
	}
	_ = m.cache.RecordNetBlock(cache.NetBlockRecord{
		Direction: "outbound", RemoteIP: firstIP, Domain: ev.Name,
		PID: ev.PID, ProcessPath: ev.Process, Matched: ind.Value, MatchedType: ind.Type,
		Level: ind.Level, Action: action,
	})
}

func (m *Manager) Health() Health {
	age := int64(0)
	stale := false
	if ls := m.lastSync.Load(); ls > 0 {
		age = time.Now().Unix() - ls
		stale = age > int64(24*3600) // >24h with no successful sync
	}
	// m.mu guards m.cfg/m.wfpOK (mutated by Reload) and the enforcer's live count
	// (mutated by applyLocked); take it so status is a coherent snapshot.
	m.mu.Lock()
	defer m.mu.Unlock()
	return Health{
		Enabled:    m.cfg.Blocklist.Enabled,
		Enforce:    m.cfg.Blocklist.Enforce,
		Indicators: m.store.Count(),
		// Real OS filter count, not the intended count: on WFP-init failure the
		// enforcer's no-op blocker reports 0, so status can't overstate protection.
		ActiveFilters:       m.enf.BlockerCount(),
		RecentBlocks:        int(m.recentBlocks.Swap(0)),
		AllowlistedDropped:  int(m.dropped.Load()),
		FeedAgeSec:          age,
		FeedStale:           stale,
		EnforcementDegraded: m.cfg.Blocklist.Enforce && !m.wfpOK,
	}
}

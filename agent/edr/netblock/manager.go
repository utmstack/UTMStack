// edr/netblock/manager.go
package netblock

import (
	"context"
	"errors"
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
	// Optional adapter injection for isolated tests; nil selects the OS backend.
	NewBlocker func(string) (Blocker, error)
}

type Health struct {
	Enabled, Enforce    bool
	Indicators          int
	ActiveFilters       int
	RecentBlocks        int
	AllowlistedDropped  int
	FeedAgeSec          int64
	FeedStale           bool
	FeedError           string
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

	lastSync          atomic.Int64 // unix sec
	recentBlocks      atomic.Int64
	dropped           atomic.Int64
	desired           atomic.Int64
	wfpOK             bool // WFP init succeeded (real enforcement possible)
	enforcementErr    error
	cleanupErr        error
	safetyErr         error
	systemNetsKey     string
	safetyEpoch       uint64
	initialReconciled bool
	// mu serializes policy and shared-filter ownership across feed reconciliation,
	// DNS events, expiry sweeps and the service's configuration reload.
	mu sync.Mutex
}

func NewManager(d Deps) *Manager {
	wfpOK := true
	makeBlocker := d.NewBlocker
	if makeBlocker == nil {
		makeBlocker = NewOSBlocker
	}
	blk, err := makeBlocker(config.ServiceName + "-Blocklist")
	if err != nil {
		logWarn("blocklist OS enforcement init failed, enforcement disabled: %v", err)
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
	// Every production sweep holds m.mu before reactive.mu. The predicate reads
	// actual enforcer ownership under that outer lock and never takes m.mu from
	// inside reactive.mu (which would invert the DNS/reload lock order).
	m.reactive.SetOnFeed(m.enf.OwnsIP)
	m.enf.retainIP = func(a netip.Addr) bool { return !m.allow.Allowed(a) && m.reactive.Owns(a) }
	m.configureSafetyLocked(d.Sys)
	return m
}

func (m *Manager) configureSafetyLocked(sys SystemNets) {
	m.safetyEpoch++
	m.systemNetsKey = systemNetsIdentity(sys)
	backend, ok := m.enf.b.(interface{ SetSafetyPolicy(SafetyPolicy) error })
	if !ok {
		return
	}
	p := SafetyPolicy{Management: sys.ServerHosts, Resolvers: sys.Resolvers, NeverBlock: append([]netip.Prefix(nil), m.allow.prefixes...)}
	for a := range m.allow.exact {
		if a.IsValid() {
			p.NeverBlock = append(p.NeverBlock, netip.PrefixFrom(a, a.BitLen()))
		}
	}
	m.safetyErr = backend.SetSafetyPolicy(p)
}

func (m *Manager) Run(ctx context.Context) {
	// A lease refresher is a backend lifecycle, not a source of packet events.
	// Wait for it to stop before closing ownership and deleting its own table.
	backendCtx, stopBackend := context.WithCancel(ctx)
	var backendWG sync.WaitGroup
	if b, ok := m.enf.b.(interface{ Run(context.Context) }); ok {
		backendWG.Add(1)
		go func() { defer backendWG.Done(); b.Run(backendCtx) }()
	}
	defer func() {
		stopBackend()
		backendWG.Wait()
		m.mu.Lock()
		defer m.mu.Unlock()
		m.cleanupErr = errors.Join(m.cleanupErr, m.enf.Clear())
		if b, ok := m.enf.b.(interface{ Close() error }); ok {
			m.cleanupErr = errors.Join(m.cleanupErr, b.Close())
		}
	}()
	// Startup convergence: re-apply from cached indicators before the first fetch.
	if recs, err := m.cache.ListIndicators(); err == nil {
		inds := make([]Indicator, 0, len(recs))
		for _, r := range recs {
			if ind, err := ParseIndicator(r.Value, r.Type, r.Level); err == nil {
				inds = append(inds, ind)
			}
		}
		m.store.Swap(inds)
		m.applyIndicators(inds)
	}

	// Always start the observers; their effects are gated by live config (read
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
				m.sweepReactive()
			}
		}
	}()

	m.feed.Run(ctx) // blocks until ctx.Done()
}

// applyIndicators recomputes the desired IP set (allowlist-filtered) and reconciles
// the enforcer. Domain/hostname indicators are ignored in Plan 1. It is the entry
// point for the feed's Run goroutine; m.mu serializes it against Reload (which runs
// on the service's reload goroutine).
func (m *Manager) applyIndicators(inds []Indicator) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastSync.Store(time.Now().Unix())
	m.applyLocked(inds)
	m.initialReconciled = true
}

// applyLocked reconciles the enforcer against the (allowlist-filtered) indicator
// set. The caller MUST hold m.mu — both applyIndicators and Reload funnel here so
// the mutex is taken exactly once per call (a non-reentrant sync.Mutex).
func (m *Manager) applyLocked(inds []Indicator) {
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
	if !m.cfg.Blocklist.Enabled || !m.cfg.Blocklist.Enforce {
		// Detect-only / enforce-off must hold no live filters — clear any that a
		// prior enforce=true cycle applied before we stop reconciling. Clear resets
		// both the exact-IP and prefix maps (and Reset removes prefix filters).
		if err := m.enf.Clear(); err != nil {
			m.enforcementErr = err
			logWarn("blocklist enforce clear: %v", err)
		} else {
			m.enforcementErr = nil
			m.cleanupErr = nil
			// Clear wiped ALL blocker filters, including reactive's shared ones — drop
			// reactive's stale expiry so a later re-enable re-blocks those IPs afresh.
			m.reactive.Reset()
		}
		m.desired.Store(0)
		return
	}
	if m.safetyErr != nil {
		m.enforcementErr = errors.Join(m.safetyErr, m.enf.Clear())
		if m.enf.BlockerCount() == 0 {
			m.reactive.Reset()
		}
		m.desired.Store(0)
		return
	}
	// Remove exemptions before and after reconciliation: either owner may have
	// held the exact filter. Successful removals relinquish stale shared claims.
	m.removeAllowedLocked()
	m.enforcementErr = m.enf.ApplyPolicy(desired, desiredP)
	if m.enforcementErr != nil {
		logWarn("blocklist enforce apply: %v", m.enforcementErr)
	}
	m.removeAllowedLocked()
	m.desired.Store(int64(len(desired)))
}

// Reload hot-applies a new config + system-nets set without a service restart: it
// rebuilds the never-block allowlist and re-derives the enforced IP set from the
// live indicator store. It runs on the service's reload goroutine while the feed
// drives applyIndicators on the manager's Run goroutine — m.mu serializes both
// around the allowlist + enforcer. Reading indicators from the cache happens before
// the lock (the cache has its own lock); the reconcile then runs under m.mu once.
func (m *Manager) Reload(cfg config.EDRConfig, sys SystemNets) {
	// Cancel and finish any old mirror transaction BEFORE taking m.mu: a feed
	// callback may already be waiting for that policy lock.
	m.feed.Reload(cfg)
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
	if current, published := m.store.snapshotPublished(); published {
		inds = snapshotIndicators(current)
	}
	m.cfg = cfg
	m.allow = BuildAllowlist(cfg, sys)
	m.configureSafetyLocked(sys)
	if m.enf.dir != cfg.Blocklist.Direction {
		if err := m.enf.Clear(); err != nil {
			m.enforcementErr = err
			return
		}
		m.enf.dir = cfg.Blocklist.Direction
		m.reactive.Reset()
		m.reactive.mu.Lock()
		m.reactive.dir = cfg.Blocklist.Direction
		m.reactive.mu.Unlock()
	}
	ttl := time.Duration(cfg.Blocklist.DomainBlockTTLMin) * time.Minute
	if ttl <= 0 {
		ttl = 30 * time.Minute
	}
	m.reactive.mu.Lock()
	m.reactive.ttl = ttl
	m.reactive.mu.Unlock()
	m.applyLocked(inds)
}

func (m *Manager) removeAllowedLocked() {
	removed, err := m.reactive.RemoveAllowed(m.allow.Allowed)
	for _, a := range removed {
		delete(m.enf.applied, a)
		delete(m.enf.uncertain, a)
	}
	if err != nil {
		logWarn("blocklist reactive exemption removal: %v", err)
	}
}

func (m *Manager) sweepReactive() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if backend, ok := m.enf.b.(interface {
		RetryCleanup() ([]netip.Addr, []netip.Prefix, error)
	}); ok {
		ips, prefixes, err := backend.RetryCleanup()
		m.cleanupErr = err
		for _, a := range ips {
			delete(m.enf.applied, a)
			delete(m.enf.uncertain, a)
		}
		for _, p := range prefixes {
			delete(m.enf.appliedP, p)
			delete(m.enf.uncertainP, p)
		}
		if err != nil {
			logWarn("blocklist partial filter cleanup: %v", err)
		}
	}
	if !m.cfg.Blocklist.Enabled || !m.cfg.Blocklist.Enforce {
		// An unsuccessful reset must not discard ownership or cleanup retries.
		if err := m.enf.Clear(); err != nil {
			m.enforcementErr = err
			logWarn("blocklist enforce clear retry: %v", err)
		} else {
			m.enforcementErr = nil
			m.cleanupErr = nil
			m.reactive.Reset()
		}
		return
	}
	if m.enforcementErr != nil {
		// Store publishes only a fully parsed and durably committed feed. Retry
		// OS-apply failures against that latest desired snapshot on the existing
		// bounded minute cadence; a malformed feed never replaces this snapshot.
		m.applyLocked(snapshotIndicators(m.store.Snapshot()))
	}
	m.removeAllowedLocked()
	if err := m.reactive.Sweep(); err != nil {
		logWarn("blocklist reactive expiry removal: %v", err)
	}
}

// onBlockEvent turns a verified EDR WFP drop into a branded event and audit
// record. The observer receives system-wide events, including drops caused by
// other filters, so raw feed membership alone cannot prove our enforcement.
func (m *Manager) onBlockEvent(be BlockEvent) {
	m.mu.Lock()
	if !m.cfg.Blocklist.Enabled || !m.cfg.Blocklist.Enforce || m.allow.Allowed(be.RemoteIP) ||
		be.Direction == "outbound" && m.enf.dir == "in" ||
		be.Direction == "inbound" && m.enf.dir == "out" ||
		!m.enf.BlocksIP(be.RemoteIP) {
		m.mu.Unlock()
		return
	}
	if b, ok := m.enf.b.(interface{ HasFilterID(uint64) bool }); ok && !b.HasFilterID(be.FilterID) {
		m.mu.Unlock()
		return
	}
	ind, ok := m.store.Match(be.RemoteIP)
	m.mu.Unlock()
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
// Manager.mu serializes shared-filter ownership and allowlist changes with the
// feed reconciler, sweeper and live configuration reload.
func (m *Manager) onDNSEvent(ev DNSEvent) {
	ind, ok := m.store.MatchName(ev.Name)
	if !ok {
		return
	}
	var toBlock []netip.Addr
	m.mu.Lock()
	allCovered := len(ev.IPs) > 0
	for _, ip := range ev.IPs {
		if !ip.IsValid() || m.allow.Allowed(ip) {
			allCovered = false
			continue
		}
		if !m.enf.BlocksIP(ip) {
			toBlock = append(toBlock, ip)
		}
	}
	// Only actually block (and report "blocked") when enforcing; in detect-only we
	// emit an honest action="detected" domain event and apply no WFP filters.
	backendHealthy := m.wfpOK && m.safetyErr == nil
	if b, ok := m.enf.b.(interface{ Health() error }); ok && b.Health() != nil {
		backendHealthy = false
	}
	enf := m.cfg.Blocklist.Enabled && m.cfg.Blocklist.Enforce && backendHealthy &&
		m.enf.dir != "in" && m.reactive.dir != "in"
	action := event.ActionDetected
	if enf && allCovered {
		if len(toBlock) == 0 || m.reactive.Block(toBlock) == len(toBlock) {
			action = event.ActionBlocked
		}
	} else if enf && len(toBlock) > 0 {
		// Mixed answers can still protect non-exempt addresses, but the event
		// remains detected because not every answer was eligible for blocking.
		m.reactive.Block(toBlock)
	}
	m.mu.Unlock()
	if action == event.ActionBlocked {
		m.recentBlocks.Add(1)
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
	feedError := ""
	if err := m.feed.HealthError(); err != nil {
		stale = true
		feedError = err.Error()
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
		FeedError:           feedError,
		EnforcementDegraded: m.enforcementHealthLocked() != nil,
	}
}

// EnforcementHealth is safe for periodic platform status: unlike Health, it
// does not consume RecentBlocks. A successful constructor alone is insufficient
// evidence if management resolution, reconciliation or lease verification fails.
func (m *Manager) EnforcementHealth() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.enforcementHealthLocked()
}

func (m *Manager) enforcementHealthLocked() error {
	var backendErr, availabilityErr error
	if m.cfg.Blocklist.Enabled && m.cfg.Blocklist.Enforce {
		if !m.wfpOK {
			availabilityErr = ErrUnsupported
		}
		if b, ok := m.enf.b.(interface{ Health() error }); ok {
			backendErr = b.Health()
		}
		availabilityErr = errors.Join(availabilityErr, m.safetyErr)
		availabilityErr = errors.Join(availabilityErr, m.initialEnforcementHealthLocked())
	}
	return errors.Join(availabilityErr, backendErr, m.enforcementErr, m.cleanupErr, m.reactive.HealthError())
}

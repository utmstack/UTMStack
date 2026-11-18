// edr/netblock/feed.go
package netblock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

// Feed pulls ThreatWinds indicators from the UTMStack mirror on a cadence and
// hot-swaps the store. applyFn runs after each successful refresh.
type Feed struct {
	cfg     config.EDRConfig
	cache   *cache.Cache
	store   *Store
	applyFn func([]Indicator)

	// cur mirrors the store's active indicator set so daily deltas can be merged
	// on top without re-fetching the full accumulative list. Guarded by mu.
	mu  sync.Mutex
	cur []Indicator

	// lastFullSync is the wall-clock time of the last successful accumulative
	// re-bootstrap (in-memory only). Zero => the first cycle does a full sync,
	// which also re-bootstraps safely after a service restart. Consulted only by
	// Run's single goroutine, so it needs no lock.
	lastFullSync  time.Time
	updateMu      sync.Mutex
	reloadMu      sync.Mutex
	cancelMu      sync.Mutex
	activeCancel  context.CancelFunc
	reloadPending atomic.Bool
	reloadCh      chan struct{}
	healthMu      sync.Mutex
	healthErr     error
	// Tests can lower limits; production always uses fixed finite upper bounds.
	limits     *networkFeedLimits
	httpClient *http.Client // socket-free HTTP transport injection in tests

	// overridable for tests
	Fetch         func(url string) (io.ReadCloser, error)
	FetchChecksum func(url string) (string, error)
}

func NewFeed(cfg config.EDRConfig, c *cache.Cache, store *Store, applyFn func([]Indicator)) *Feed {
	f := &Feed{cfg: cloneFeedConfig(cfg), cache: c, store: store, applyFn: applyFn, reloadCh: make(chan struct{}, 1)}
	// Seed cur from any indicators the store already holds so a delta applied
	// before the first accumulative bootstrap still merges onto the live set.
	f.cur = snapshotIndicators(store.Snapshot())
	return f
}

func cloneFeedConfig(cfg config.EDRConfig) config.EDRConfig {
	cfg.Blocklist.Levels = append([]int(nil), cfg.Blocklist.Levels...)
	cfg.Blocklist.IndicatorTypes = append([]string(nil), cfg.Blocklist.IndicatorTypes...)
	return cfg
}

// Reload cancels an old-source request before publishing the new configuration.
// It serializes with the complete fetch/verify/persist/apply transaction, so a
// completed reload cannot later be overwritten by an old mirror's response.
// Callers must not hold the manager's policy lock while waiting here.
func (f *Feed) Reload(cfg config.EDRConfig) {
	f.reloadMu.Lock()
	defer f.reloadMu.Unlock()
	f.reloadPending.Store(true)
	f.cancelMu.Lock()
	if f.activeCancel != nil {
		f.activeCancel()
	}
	f.cancelMu.Unlock()
	f.updateMu.Lock()
	f.cfg = cloneFeedConfig(cfg)
	f.lastFullSync = time.Time{}
	f.reloadPending.Store(false)
	f.updateMu.Unlock()
	select {
	case f.reloadCh <- struct{}{}:
	default:
	}
}

func (f *Feed) updateContext(parent context.Context) (context.Context, func()) {
	ctx, cancel := context.WithCancel(parent)
	f.cancelMu.Lock()
	f.activeCancel = cancel
	if f.reloadPending.Load() {
		cancel()
	}
	f.cancelMu.Unlock()
	return ctx, func() { cancel(); f.cancelMu.Lock(); f.activeCancel = nil; f.cancelMu.Unlock() }
}

func boundedFeedHours(hours int, fallback time.Duration) time.Duration {
	if hours <= 0 || hours > 24*365 {
		return fallback
	}
	return time.Duration(hours) * time.Hour
}

func (f *Feed) refreshInterval() time.Duration {
	f.updateMu.Lock()
	defer f.updateMu.Unlock()
	return boundedFeedHours(f.cfg.Blocklist.RefreshHours, 6*time.Hour)
}

// snapshotIndicators enumerates the indicators held by a Matcher. It reaches into
// the Matcher's package-private fields (same package) because the store exposes no
// public enumerator; feed.go is the only consumer that needs the reverse mapping.
func snapshotIndicators(m *Matcher) []Indicator {
	if m == nil {
		return nil
	}
	out := make([]Indicator, 0, m.Len())
	for _, ind := range m.exact {
		out = append(out, ind)
	}
	for _, pe := range m.prefixes {
		out = append(out, pe.ind)
	}
	out = append(out, m.nameInd...)
	return out
}

func (f *Feed) baseURL() string {
	if f.cfg.Blocklist.MirrorBaseURL != "" {
		return strings.TrimRight(f.cfg.Blocklist.MirrorBaseURL, "/")
	}
	if f.cfg.Server == "" {
		return ""
	}
	return "https://" + f.cfg.Server + ":" + config.MirrorPort + config.MirrorBasePath + "/feeds/v1"
}

func (f *Feed) skipTLS() bool { return f.cfg.SkipCertValidate }

func (f *Feed) client() *http.Client {
	if f.httpClient != nil {
		return f.httpClient
	}
	c := &http.Client{Timeout: 5 * time.Minute}
	if f.skipTLS() {
		c.Transport = &http.Transport{TLSClientConfig: insecureTLS()}
	}
	return c
}

func (f *Feed) httpFetch(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := f.client().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	if resp.ContentLength > MaxNetworkFeedRawBytes {
		resp.Body.Close()
		return nil, ErrNetworkFeedBudget
	}
	return resp.Body, nil
}

func (f *Feed) httpChecksum(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", err
	}
	resp, err := f.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum %s: status %d", url, resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4097))
	if len(b) > 4096 {
		return "", ErrNetworkFeedBudget
	}
	return strings.TrimSpace(string(b)), err
}

func (f *Feed) fetch(ctx context.Context, url string) (io.ReadCloser, error) {
	if f.Fetch != nil {
		return f.Fetch(url)
	}
	return f.httpFetch(ctx, url)
}
func (f *Feed) checksum(ctx context.Context, url string) (string, error) {
	if f.FetchChecksum != nil {
		return f.FetchChecksum(url)
	}
	return f.httpChecksum(ctx, url)
}
func (f *Feed) HealthError() error     { f.healthMu.Lock(); defer f.healthMu.Unlock(); return f.healthErr }
func (f *Feed) recordHealth(err error) { f.healthMu.Lock(); f.healthErr = err; f.healthMu.Unlock() }
func (f *Feed) budgets() (networkFeedLimits, error) {
	l := defaultNetworkFeedLimits()
	if f.limits != nil {
		l = *f.limits
	}
	if !l.valid() {
		return l, ErrNetworkFeedBudget
	}
	return l, nil
}
func feedArtifactCount(types []string, levels []int, max int) error {
	if len(types) == 0 || len(levels) == 0 || len(types) > max || len(levels) > max || len(types)*len(levels) > max {
		return ErrNetworkFeedBudget
	}
	for _, typ := range types {
		if typ != "ip" && typ != "domain" && typ != "hostname" {
			return errors.New("unsupported network feed artifact type")
		}
	}
	for _, level := range levels {
		if level < 1 || level > 3 {
			return errors.New("unsupported network feed level")
		}
	}
	return nil
}

// Run drives the immediate-then-ticker refresh loop. Each
// cycle chooses between a full accumulative re-sync and a lightweight daily
// delta: a full sync runs on the first cycle (lastFullSync zero) and thereafter
// every FullResyncHours; intervening cycles apply only the daily deltas. This
// keeps the periodic-full / daily-delta cadence the mirror publishes, while the
// first cycle always re-bootstraps safely after a restart.
func (f *Feed) Run(ctx context.Context) {
	cycle := func() {
		f.updateMu.Lock()
		enabled := f.cfg.Blocklist.Enabled
		full := boundedFeedHours(f.cfg.Blocklist.FullResyncHours, 24*time.Hour)
		needsFull := f.lastFullSync.IsZero() || time.Since(f.lastFullSync) >= full
		f.updateMu.Unlock()
		if !enabled {
			return
		}
		if needsFull {
			if err := f.UpdateOnce(ctx); err != nil {
				logWarn("blocklist full sync failed: %v", err)
			}
		} else {
			if err := f.updateDeltaOnce(ctx); err != nil {
				logWarn("blocklist daily delta failed: %v", err)
			}
		}
	}
	cycle()
	t := time.NewTicker(f.refreshInterval())
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cycle()
		case <-f.reloadCh:
			t.Reset(f.refreshInterval())
			cycle()
		}
	}
}

// UpdateOnce pulls the accumulative snapshot for every configured (type, level)
// feed, verifies checksum, rebuilds the store, and persists indicators. Every
// configured artifact must be fetched, verified and parsed before publishing
// any of this cycle. A partial mirror failure must not silently remove the
// previous artifact's indicators or report a partial generation as fresh.
func (f *Feed) UpdateOnce(ctx context.Context) (retErr error) {
	f.updateMu.Lock()
	defer f.updateMu.Unlock()
	defer func() { f.recordHealth(retErr) }()
	ctx, finish := f.updateContext(ctx)
	defer finish()
	limits, err := f.budgets()
	if err != nil {
		return err
	}
	rawRemaining, expandedRemaining := limits.RawTotal, limits.ExpandedTotal
	base := f.baseURL()
	if base == "" {
		return fmt.Errorf("blocklist mirror not configured (no server registered)")
	}
	var all []Indicator
	success := 0 // artifacts fully fetched+verified+parsed
	seen := map[string]struct{}{}
	types := f.cfg.Blocklist.IndicatorTypes
	if len(types) == 0 {
		types = []string{"ip"}
	}
	if err := feedArtifactCount(types, f.cfg.Blocklist.Levels, limits.Artifacts); err != nil {
		return err
	}
	var states []cache.FeedState
	for _, typ := range types {
		for _, lvl := range f.cfg.Blocklist.Levels {
			level := fmt.Sprintf("level%d", lvl)
			url := fmt.Sprintf("%s/download/list/%s/accumulative/%s", base, level, typ)
			ckURL := fmt.Sprintf("%s/download/list/%s/accumulative/%s.sha256", base, level, typ)

			if err := ctx.Err(); err != nil {
				return err
			}
			body, err := f.fetch(ctx, url)
			if err != nil {
				return fmt.Errorf("blocklist fetch %s: %w", url, err)
			}
			raw, err := readNetworkFeed(body, smallerBudget(limits.RawArtifact, rawRemaining))
			body.Close()
			if err != nil {
				return fmt.Errorf("blocklist read %s: %w", url, err)
			}
			rawRemaining -= int64(len(raw))
			want, err := f.checksum(ctx, ckURL)
			if err != nil {
				return fmt.Errorf("blocklist checksum %s: %w", ckURL, err)
			}
			if want != "" {
				got := sha256.Sum256(raw)
				if hex.EncodeToString(got[:]) != want {
					return fmt.Errorf("blocklist checksum mismatch for %s; keeping last-known-good set", url)
				}
			} else {
				return errors.New("network feed checksum is empty")
			}
			inds, consumed, err := parseAccumulativeBounded(bytesReader(raw), lvl, typ, smallerBudget(limits.ExpandedArtifact, expandedRemaining), limits.Indicators)
			expandedRemaining -= consumed
			if err != nil {
				return fmt.Errorf("blocklist parse %s: %w", url, err)
			}
			for _, ind := range inds {
				key := ind.Type + "|" + ind.Value
				if _, ok := seen[key]; ok {
					continue
				}
				if len(all) >= limits.Indicators {
					return ErrNetworkFeedBudget
				}
				seen[key] = struct{}{}
				all = append(all, ind)
			}
			success++
			states = append(states, cache.FeedState{
				Feed: level + "/" + typ, LastFullSync: time.Now().UTC(),
				Checksum: want, IndicatorNum: len(inds),
			})
		}
	}
	// No artifacts selected is not an authoritative empty snapshot. A complete
	// cycle whose configured artifacts are all verified and genuinely empty is
	// valid and may remove old indicators, including from durable storage.
	if err := ctx.Err(); err != nil {
		return err
	}
	if success == 0 {
		return fmt.Errorf("blocklist: all feed artifacts failed; keeping last-known-good set")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := f.persist(all, states...); err != nil {
		return err
	}
	f.store.Swap(all)
	f.mu.Lock()
	f.cur = all
	f.mu.Unlock()
	if f.applyFn != nil {
		f.applyFn(all)
	}
	f.lastFullSync = time.Now()
	return nil
}

// updateDeltaOnce pulls the daily incremental delta for every configured
// (type, level) feed and layers it onto the last-known-good accumulative set. The
// daily artifact is PLAIN (uncompressed) ndjson — one {"value","type","op"} per
// line — with a bare-hex .sha256 sibling over the ndjson bytes (unlike the gzip'd
// accumulative snapshot). The configured cycle is atomic: any fetch, read,
// checksum or parser failure leaves every prior indicator and freshness state
// intact. A complete verified empty cycle is a valid no-op.
func (f *Feed) updateDeltaOnce(ctx context.Context) (retErr error) {
	f.updateMu.Lock()
	defer f.updateMu.Unlock()
	defer func() { f.recordHealth(retErr) }()
	ctx, finish := f.updateContext(ctx)
	defer finish()
	limits, err := f.budgets()
	if err != nil {
		return err
	}
	rawRemaining, expandedRemaining := limits.RawTotal, limits.ExpandedTotal
	base := f.baseURL()
	if base == "" {
		return fmt.Errorf("blocklist mirror not configured (no server registered)")
	}
	var ops []DailyOp
	success := 0 // daily artifacts fully fetched + verified + parsed
	types := f.cfg.Blocklist.IndicatorTypes
	if len(types) == 0 {
		types = []string{"ip"}
	}
	if err := feedArtifactCount(types, f.cfg.Blocklist.Levels, limits.Artifacts); err != nil {
		return err
	}
	for _, typ := range types {
		for _, lvl := range f.cfg.Blocklist.Levels {
			level := fmt.Sprintf("level%d", lvl)
			url := fmt.Sprintf("%s/download/list/%s/daily/%s", base, level, typ)
			ckURL := url + ".sha256"

			if err := ctx.Err(); err != nil {
				return err
			}
			body, err := f.fetch(ctx, url)
			if err != nil {
				return fmt.Errorf("blocklist daily fetch %s: %w", url, err)
			}
			raw, err := readNetworkFeed(body, smallerBudget(limits.RawArtifact, rawRemaining))
			body.Close()
			if err != nil {
				return fmt.Errorf("blocklist daily read %s: %w", url, err)
			}
			rawRemaining -= int64(len(raw))
			want, err := f.checksum(ctx, ckURL)
			if err != nil {
				return fmt.Errorf("blocklist daily checksum %s: %w", ckURL, err)
			}
			if want != "" {
				got := sha256.Sum256(raw)
				if hex.EncodeToString(got[:]) != want {
					return fmt.Errorf("blocklist daily checksum mismatch for %s; keeping last-known-good set", url)
				}
			} else {
				return errors.New("network feed checksum is empty")
			}
			// Daily is PLAIN ndjson — do NOT gunzip (ParseDaily reads it directly).
			delta, consumed, err := parseDailyBounded(bytesReader(raw), smallerBudget(limits.ExpandedArtifact, expandedRemaining), limits.DailyOps)
			expandedRemaining -= consumed
			if err != nil {
				return fmt.Errorf("blocklist daily parse %s: %w", url, err)
			}
			if len(delta) > limits.DailyOps-len(ops) {
				return ErrNetworkFeedBudget
			}
			ops = append(ops, delta...)
			success++
		}
	}
	// No configured artifacts is not evidence of a successful refresh.
	if err := ctx.Err(); err != nil {
		return err
	}
	if success == 0 {
		return fmt.Errorf("blocklist: all daily artifacts failed; keeping last-known-good set")
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if f.cache == nil {
		return errors.New("network indicator cache is unavailable")
	}
	if err := f.applyDelta(ops); err != nil {
		return err
	}
	return nil
}

// applyDelta merges add/del daily ops into the current in-memory indicator set and
// hot-swaps the store. Deltas layer on top of the accumulative bootstrap held in
// f.cur, so a daily fetch failure leaves the last-good set intact.
func (f *Feed) applyDelta(ops []DailyOp) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	limits, err := f.budgets()
	if err != nil {
		return err
	}
	if len(ops) > limits.DailyOps || len(f.cur) > limits.Indicators {
		return ErrNetworkFeedBudget
	}
	idx := map[string]Indicator{} // key = type|value
	for _, ind := range f.cur {
		idx[ind.Type+"|"+ind.Value] = ind
	}
	// Apply ALL dels before ANY add. The server diffs raw upstream strings while
	// indicators are keyed by their canonical form (domains lowercased, CIDRs
	// masked), so a representation change (e.g. Evil.com → evil.com) arrives as a
	// del of the old spelling plus an add of the new one that map to the SAME
	// key. Dels-first makes the add win, keeping the still-listed indicator —
	// the correct outcome, and independent of the ops' arrival order.
	for _, op := range ops {
		if op.Op != "del" {
			continue
		}
		if ind, err := ParseIndicator(op.Value, op.Type, 0); err == nil {
			delete(idx, ind.Type+"|"+ind.Value)
		}
	}
	for _, op := range ops {
		if op.Op != "add" {
			continue
		}
		if ind, err := ParseIndicator(op.Value, op.Type, 0); err == nil {
			key := ind.Type + "|" + ind.Value
			if _, exists := idx[key]; !exists && len(idx) >= limits.Indicators {
				return ErrNetworkFeedBudget
			}
			idx[key] = ind
		}
	}
	next := make([]Indicator, 0, len(idx))
	for _, ind := range idx {
		next = append(next, ind)
	}
	if f.cache != nil {
		if err := f.persist(next); err != nil {
			return err
		}
	}
	f.cur = next
	f.store.Swap(next)
	if f.applyFn != nil {
		f.applyFn(next)
	}
	return nil
}

func (f *Feed) persist(inds []Indicator, states ...cache.FeedState) error {
	if f.cache == nil {
		return errors.New("network indicator cache is unavailable")
	}
	now := time.Now().UTC()
	recs := make([]cache.IndicatorRecord, 0, len(inds))
	for _, ind := range inds {
		recs = append(recs, cache.IndicatorRecord{
			Value: ind.Value, Type: ind.Type, Level: ind.Level,
			ListName: fmt.Sprintf("level%d/%s", ind.Level, ind.Type), FirstSeen: now, LastSeen: now,
		})
	}
	if err := f.cache.ReplaceIndicators(recs, states...); err != nil {
		return fmt.Errorf("blocklist persist snapshot: %w", err)
	}
	return nil
}

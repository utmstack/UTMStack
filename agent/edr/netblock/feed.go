// edr/netblock/feed.go
package netblock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
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
	lastFullSync time.Time

	// overridable for tests
	Fetch         func(url string) (io.ReadCloser, error)
	FetchChecksum func(url string) (string, error)
}

func NewFeed(cfg config.EDRConfig, c *cache.Cache, store *Store, applyFn func([]Indicator)) *Feed {
	f := &Feed{cfg: cfg, cache: c, store: store, applyFn: applyFn}
	f.Fetch = f.httpFetch
	f.FetchChecksum = f.httpChecksum
	// Seed cur from any indicators the store already holds so a delta applied
	// before the first accumulative bootstrap still merges onto the live set.
	f.cur = snapshotIndicators(store.Snapshot())
	return f
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
	c := &http.Client{Timeout: 5 * time.Minute}
	if f.skipTLS() {
		c.Transport = &http.Transport{TLSClientConfig: insecureTLS()}
	}
	return c
}

func (f *Feed) httpFetch(url string) (io.ReadCloser, error) {
	resp, err := f.client().Get(url)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	return resp.Body, nil
}

func (f *Feed) httpChecksum(url string) (string, error) {
	resp, err := f.client().Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("checksum %s: status %d", url, resp.StatusCode)
	}
	b, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return strings.TrimSpace(string(b)), err
}

// Run drives the immediate-then-ticker refresh loop (freshclam pattern). Each
// cycle chooses between a full accumulative re-sync and a lightweight daily
// delta: a full sync runs on the first cycle (lastFullSync zero) and thereafter
// every FullResyncHours; intervening cycles apply only the daily deltas. This
// keeps the periodic-full / daily-delta cadence the mirror publishes, while the
// first cycle always re-bootstraps safely after a restart.
func (f *Feed) Run(ctx context.Context) {
	full := time.Duration(f.cfg.Blocklist.FullResyncHours) * time.Hour
	if full <= 0 {
		full = 24 * time.Hour
	}
	cycle := func() {
		if f.lastFullSync.IsZero() || time.Since(f.lastFullSync) >= full {
			if err := f.UpdateOnce(ctx); err != nil {
				logWarn("blocklist full sync failed: %v", err)
			} else {
				f.lastFullSync = time.Now()
			}
		} else {
			if err := f.updateDeltaOnce(ctx); err != nil {
				logWarn("blocklist daily delta failed: %v", err)
			}
		}
	}
	cycle()
	hrs := f.cfg.Blocklist.RefreshHours
	if hrs <= 0 {
		hrs = 6
	}
	t := time.NewTicker(time.Duration(hrs) * time.Hour)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			cycle()
		}
	}
}

// UpdateOnce pulls the accumulative snapshot for every configured (type, level)
// feed, verifies checksum, rebuilds the store, and persists indicators. Each
// (type, level) artifact is fetched best-effort: a fetch/parse failure or a
// checksum mismatch skips only that artifact (log + continue) so a mirror that
// does not serve one indicator type — or a single corrupt artifact — never
// aborts the whole update or swaps in unverified data. (Daily deltas layer on in
// Task N via UpdateOnce reuse; Plan 1 bootstraps full each cycle for the modest
// list — deltas are a later optimization.)
func (f *Feed) UpdateOnce(ctx context.Context) error {
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
	for _, typ := range types {
		for _, lvl := range f.cfg.Blocklist.Levels {
			level := fmt.Sprintf("level%d", lvl)
			url := fmt.Sprintf("%s/download/list/%s/accumulative/%s", base, level, typ)
			ckURL := fmt.Sprintf("%s/download/list/%s/accumulative/%s.sha256", base, level, typ)

			body, err := f.Fetch(url)
			if err != nil {
				logWarn("blocklist fetch %s failed: %v", url, err)
				continue
			}
			raw, err := io.ReadAll(body)
			body.Close()
			if err != nil {
				logWarn("blocklist read %s failed: %v", url, err)
				continue
			}
			want, err := f.FetchChecksum(ckURL)
			if err != nil {
				logWarn("blocklist checksum %s failed: %v", ckURL, err)
				continue
			}
			if want != "" {
				got := sha256.Sum256(raw)
				if hex.EncodeToString(got[:]) != want {
					// Never swap in unverified data; skip this artifact only.
					logWarn("blocklist checksum mismatch for %s; skipping artifact", url)
					continue
				}
			}
			inds, err := ParseAccumulativeTyped(bytesReader(raw), lvl, typ)
			if err != nil {
				logWarn("blocklist parse %s failed: %v", url, err)
				continue
			}
			for _, ind := range inds {
				key := ind.Type + "|" + ind.Value
				if _, ok := seen[key]; ok {
					continue
				}
				seen[key] = struct{}{}
				all = append(all, ind)
			}
			success++
			_ = f.cache.SaveFeedState(cache.FeedState{
				Feed: level + "/" + typ, LastFullSync: time.Now().UTC(),
				Checksum: want, IndicatorNum: len(inds),
			})
		}
	}
	// Stale-never-empty: if every artifact failed (mirror outage, checksum
	// mismatch, parse error), keep the last-known-good set. Do NOT swap in the
	// empty set, do NOT persist, and do NOT run applyFn (which would clear the
	// live filters and advance lastSync — falsely reporting the feed fresh).
	// A genuinely empty but successfully-fetched artifact (success>0) is still
	// allowed to swap; only TOTAL failure is guarded.
	if success == 0 {
		return fmt.Errorf("blocklist: all feed artifacts failed; keeping last-known-good set")
	}
	f.store.Swap(all)
	f.mu.Lock()
	f.cur = all
	f.mu.Unlock()
	f.persist(all)
	if f.applyFn != nil {
		f.applyFn(all)
	}
	return nil
}

// updateDeltaOnce pulls the daily incremental delta for every configured
// (type, level) feed and layers it onto the last-known-good accumulative set. The
// daily artifact is PLAIN (uncompressed) ndjson — one {"value","type","op"} per
// line — with a bare-hex .sha256 sibling over the ndjson bytes (unlike the gzip'd
// accumulative snapshot). Each artifact is best-effort: a fetch/read/checksum
// failure or a checksum mismatch skips only that artifact (log + continue). If
// EVERY daily artifact fails (success == 0) the delta is NOT applied and an error
// is returned so the live set is never wiped (mirrors UpdateOnce's Finding-B
// guard). A genuinely empty but successfully-fetched delta (success > 0, no ops)
// is a valid no-op.
func (f *Feed) updateDeltaOnce(ctx context.Context) error {
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
	for _, typ := range types {
		for _, lvl := range f.cfg.Blocklist.Levels {
			level := fmt.Sprintf("level%d", lvl)
			url := fmt.Sprintf("%s/download/list/%s/daily/%s", base, level, typ)
			ckURL := url + ".sha256"

			body, err := f.Fetch(url)
			if err != nil {
				logWarn("blocklist daily fetch %s failed: %v", url, err)
				continue
			}
			raw, err := io.ReadAll(body)
			body.Close()
			if err != nil {
				logWarn("blocklist daily read %s failed: %v", url, err)
				continue
			}
			want, err := f.FetchChecksum(ckURL)
			if err != nil {
				logWarn("blocklist daily checksum %s failed: %v", ckURL, err)
				continue
			}
			if want != "" {
				got := sha256.Sum256(raw)
				if hex.EncodeToString(got[:]) != want {
					// Never apply unverified data; skip this artifact only.
					logWarn("blocklist daily checksum mismatch for %s; skipping artifact", url)
					continue
				}
			}
			// Daily is PLAIN ndjson — do NOT gunzip (ParseDaily reads it directly).
			delta, _ := ParseDaily(bytesReader(raw))
			ops = append(ops, delta...)
			success++
		}
	}
	// Stale-never-empty: if every daily artifact failed, keep the last-known-good
	// set — do NOT apply and do NOT persist.
	if success == 0 {
		return fmt.Errorf("blocklist: all daily artifacts failed; keeping last-known-good set")
	}
	f.applyDelta(ops)
	// Persist the merged set applyDelta produced (read f.cur under mu, matching
	// applyDelta's own locking discipline).
	f.mu.Lock()
	merged := f.cur
	f.mu.Unlock()
	f.persist(merged)
	return nil
}

// applyDelta merges add/del daily ops into the current in-memory indicator set and
// hot-swaps the store. Deltas layer on top of the accumulative bootstrap held in
// f.cur, so a daily fetch failure leaves the last-good set intact.
func (f *Feed) applyDelta(ops []DailyOp) {
	f.mu.Lock()
	defer f.mu.Unlock()
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
			idx[ind.Type+"|"+ind.Value] = ind
		}
	}
	next := make([]Indicator, 0, len(idx))
	for _, ind := range idx {
		next = append(next, ind)
	}
	f.cur = next
	f.store.Swap(next)
	if f.applyFn != nil {
		f.applyFn(next)
	}
}

func (f *Feed) persist(inds []Indicator) {
	now := time.Now().UTC()
	recs := make([]cache.IndicatorRecord, 0, len(inds))
	for _, ind := range inds {
		recs = append(recs, cache.IndicatorRecord{
			Value: ind.Value, Type: ind.Type, Level: ind.Level,
			ListName: fmt.Sprintf("level%d/%s", ind.Level, ind.Type), FirstSeen: now, LastSeen: now,
		})
	}
	if err := f.cache.PutIndicators(recs); err != nil {
		logWarn("blocklist persist indicators: %v", err)
	}
}

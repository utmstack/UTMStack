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

	// overridable for tests
	Fetch         func(url string) (io.ReadCloser, error)
	FetchChecksum func(url string) (string, error)
}

func NewFeed(cfg config.EDRConfig, c *cache.Cache, store *Store, applyFn func([]Indicator)) *Feed {
	f := &Feed{cfg: cfg, cache: c, store: store, applyFn: applyFn}
	f.Fetch = f.httpFetch
	f.FetchChecksum = f.httpChecksum
	return f
}

func (f *Feed) baseURL() string {
	if f.cfg.Blocklist.MirrorBaseURL != "" {
		return strings.TrimRight(f.cfg.Blocklist.MirrorBaseURL, "/")
	}
	return strings.TrimRight(f.cfg.Server, "/") + "/feeds/v1"
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

// Run drives the immediate-then-ticker refresh loop (freshclam pattern).
func (f *Feed) Run(ctx context.Context) {
	if err := f.UpdateOnce(ctx); err != nil {
		logWarn("blocklist initial feed update failed: %v", err)
	}
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
			if err := f.UpdateOnce(ctx); err != nil {
				logWarn("blocklist feed update failed: %v", err)
			}
		}
	}
}

// UpdateOnce pulls the accumulative snapshot for every configured (level, ip)
// feed, verifies checksum, rebuilds the store, and persists indicators. (Daily
// deltas layer on in Task N via UpdateOnce reuse; Plan 1 bootstraps full each cycle
// for the modest IP list — deltas are a later optimization.)
func (f *Feed) UpdateOnce(ctx context.Context) error {
	var all []Indicator
	seen := map[string]struct{}{}
	for _, lvl := range f.cfg.Blocklist.Levels {
		level := fmt.Sprintf("level%d", lvl)
		url := fmt.Sprintf("%s/download/list/%s/accumulative/ip", f.baseURL(), level)
		ckURL := fmt.Sprintf("%s/download/checksum?level=%s&type=accumulative&name=ip", f.baseURL(), level)

		body, err := f.Fetch(url)
		if err != nil {
			return err
		}
		raw, err := io.ReadAll(body)
		body.Close()
		if err != nil {
			return err
		}
		want, err := f.FetchChecksum(ckURL)
		if err != nil {
			return err
		}
		if want != "" {
			got := sha256.Sum256(raw)
			if hex.EncodeToString(got[:]) != want {
				return fmt.Errorf("checksum mismatch for %s", url)
			}
		}
		inds, err := ParseAccumulative(bytesReader(raw), lvl)
		if err != nil {
			return err
		}
		for _, ind := range inds {
			key := ind.Type + "|" + ind.Value
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
			all = append(all, ind)
		}
		_ = f.cache.SaveFeedState(cache.FeedState{
			Feed: level + "/ip", LastFullSync: time.Now().UTC(),
			Checksum: want, IndicatorNum: len(inds),
		})
	}
	f.store.Swap(all)
	f.persist(all)
	if f.applyFn != nil {
		f.applyFn(all)
	}
	return nil
}

func (f *Feed) persist(inds []Indicator) {
	now := time.Now().UTC()
	recs := make([]cache.IndicatorRecord, 0, len(inds))
	for _, ind := range inds {
		recs = append(recs, cache.IndicatorRecord{
			Value: ind.Value, Type: ind.Type, Level: ind.Level,
			ListName: fmt.Sprintf("level%d/ip", ind.Level), FirstSeen: now, LastSeen: now,
		})
	}
	if err := f.cache.PutIndicators(recs); err != nil {
		logWarn("blocklist persist indicators: %v", err)
	}
}

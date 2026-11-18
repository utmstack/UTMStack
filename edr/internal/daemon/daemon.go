// edr/internal/daemon/daemon.go
// Package daemon wires the mirror producers: a cvdupdate loop for signature
// databases and a ThreatWinds sync loop for indicator feeds, both writing
// into the shared mirror volume that agentmanager serves to the fleet.
package daemon

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/utmstack/UTMStack/edr/internal/feeds"
	"github.com/utmstack/UTMStack/edr/internal/mirror"
	"github.com/utmstack/UTMStack/edr/internal/serve"
	"github.com/utmstack/UTMStack/edr/internal/sigs"
)

type Config struct {
	MirrorDir   string
	HTTPPort    string
	SigEvery    time.Duration
	FeedEvery   time.Duration
	Levels      []int
	Names       []string
	TWURL       string
	TWKey       string
	TWSecret    string
	UTMHost     string
	InternalKey string
}

func FromEnv(getenv func(string) string) (Config, error) {
	cfg := Config{
		MirrorDir:   "/mirror",
		HTTPPort:    "9002",
		SigEvery:    time.Hour,
		FeedEvery:   6 * time.Hour,
		Levels:      []int{1},
		Names:       []string{"ip", "domain", "hostname"},
		TWURL:       "https://apis.threatwinds.com",
		UTMHost:     "http://backend:8080",
		TWKey:       getenv("TW_API_KEY"),
		TWSecret:    getenv("TW_API_SECRET"),
		InternalKey: getenv("INTERNAL_KEY"),
	}
	if v := getenv("EDR_MIRROR_DIR"); v != "" {
		cfg.MirrorDir = v
	}
	if v := getenv("EDR_HTTP_PORT"); v != "" {
		cfg.HTTPPort = v
	}
	if v := getenv("EDR_SIG_SYNC_MINUTES"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return cfg, fmt.Errorf("EDR_SIG_SYNC_MINUTES: %q", v)
		}
		cfg.SigEvery = time.Duration(n) * time.Minute
	}
	if v := getenv("EDR_FEED_SYNC_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return cfg, fmt.Errorf("EDR_FEED_SYNC_HOURS: %q", v)
		}
		cfg.FeedEvery = time.Duration(n) * time.Hour
	}
	if v := getenv("EDR_FEED_LEVELS"); v != "" {
		cfg.Levels = nil
		for _, part := range strings.Split(v, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(part))
			if err != nil {
				return cfg, fmt.Errorf("EDR_FEED_LEVELS: %q", v)
			}
			cfg.Levels = append(cfg.Levels, n)
		}
	}
	if v := getenv("EDR_FEED_NAMES"); v != "" {
		cfg.Names = nil
		for _, part := range strings.Split(v, ",") {
			cfg.Names = append(cfg.Names, strings.TrimSpace(part))
		}
	}
	if v := getenv("TW_API_URL"); v != "" {
		cfg.TWURL = v
	}
	if v := getenv("UTM_HOST"); v != "" {
		cfg.UTMHost = v
	}
	if cfg.InternalKey == "" && (cfg.TWKey == "" || cfg.TWSecret == "") {
		return cfg, errors.New("INTERNAL_KEY (or TW_API_KEY/TW_API_SECRET) is required")
	}
	return cfg, nil
}

func Run(ctx context.Context, cfg Config) error {
	rec := mirror.NewRecorder(cfg.MirrorDir)

	// Plain-HTTP mirror surface for self-signed deployments.
	go func() {
		if err := serve.Serve(ctx, ":"+cfg.HTTPPort, cfg.MirrorDir); err != nil {
			log.Printf("http serve: %v", err)
		}
	}()

	// Signature loop.
	sig := sigs.New(cfg.MirrorDir+"/signatures", rec)
	if err := sig.EnsureConfig(); err != nil {
		return err
	}
	go loop(ctx, "signatures", cfg.SigEvery, func() {
		if err := sig.SyncOnce(); err != nil {
			log.Printf("signature sync: %v", err)
		}
	})

	// Feed loop (resolve credentials first; the backend owns registration).
	go func() {
		key, secret := cfg.TWKey, cfg.TWSecret
		for key == "" || secret == "" {
			var err error
			key, secret, err = feeds.FetchCredentials(ctx, cfg.UTMHost, cfg.InternalKey, http.DefaultClient)
			if err == nil && key != "" && secret != "" {
				break
			}
			log.Printf("threat-intel credentials not available yet: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Minute):
			}
		}
		client := &feeds.Client{BaseURL: cfg.TWURL, APIKey: key, APISecret: secret,
			HTTP: &http.Client{Timeout: 5 * time.Minute}}
		syncer := &feeds.Syncer{Root: cfg.MirrorDir, Levels: cfg.Levels, Names: cfg.Names,
			Fetch: client.FetchAccumulative, Rec: rec}
		loop(ctx, "feeds", cfg.FeedEvery, func() { syncer.SyncOnce(ctx) })
	}()

	<-ctx.Done()
	return nil
}

// loop runs fn immediately, then on every tick with up to 10% jitter so a
// fleet of servers doesn't hit upstreams in lockstep.
func loop(ctx context.Context, name string, every time.Duration, fn func()) {
	fn()
	for {
		jitter := time.Duration(rand.Int63n(int64(every / 10)))
		select {
		case <-ctx.Done():
			return
		case <-time.After(every + jitter):
			fn()
		}
	}
}

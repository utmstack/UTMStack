package feed

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestFailoverToCDNAfterRepeatedMirrorFailures(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.SigMirror = "https://m.example.com/sigs"
	cfg.SignatureFallback = "cdn"

	f := New(cfg, c)
	var wroteCDN, restoredMirror bool
	f.writeConf = func(useCDN bool) error {
		if useCDN {
			wroteCDN = true
		} else if wroteCDN {
			restoredMirror = true
		}
		return nil
	}
	f.runUpdater = func() error {
		if wroteCDN {
			return nil // CDN cycle succeeds
		}
		return errors.New("mirror unreachable")
	}
	f.currentSigDB = func() (string, error) { return "300", nil }

	for i := 0; i < mirrorFailThreshold; i++ {
		_, _ = f.updateOnce()
	}
	if !wroteCDN {
		t.Fatalf("expected CDN failover after %d mirror failures", mirrorFailThreshold)
	}
	if !restoredMirror {
		t.Fatal("mirror config must be restored after the successful CDN cycle")
	}
	if f.mirrorFails != 0 || f.onCDNFallback {
		t.Fatalf("state not reset: fails=%d onCDN=%v", f.mirrorFails, f.onCDNFallback)
	}
}

func TestNoFailoverWhenPolicyNone(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.SigMirror = "https://m.example.com/sigs"
	cfg.SignatureFallback = "none"

	f := New(cfg, c)
	var wroteCDN bool
	f.writeConf = func(useCDN bool) error { wroteCDN = wroteCDN || useCDN; return nil }
	f.runUpdater = func() error { return errors.New("mirror unreachable") }
	for i := 0; i < mirrorFailThreshold+2; i++ {
		_, _ = f.updateOnce()
	}
	if wroteCDN {
		t.Fatal("policy none must never fall back to the CDN")
	}
}

func TestNextDelayRetriesFasterAfterFailure(t *testing.T) {
	interval := 4 * time.Hour
	if d := nextDelay(nil, interval); d != interval {
		t.Fatalf("success delay = %v", d)
	}
	if d := nextDelay(errors.New("x"), interval); d != 15*time.Minute {
		t.Fatalf("failure delay = %v", d)
	}
	if d := nextDelay(errors.New("x"), 5*time.Minute); d != 5*time.Minute {
		t.Fatalf("failure delay must be capped at the interval, got %v", d)
	}
}

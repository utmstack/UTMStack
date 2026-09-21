package feed

import (
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestSignatureSource(t *testing.T) {
	cfg := config.Default()
	if SignatureSource(cfg) != "official-cdn" {
		t.Fatal("no server registered → official-cdn")
	}
	cfg.Server = "utm.example.com"
	if SignatureSource(cfg) != "utmstack-mirror" {
		t.Fatal("auto-derived mirror → utmstack-mirror")
	}
	cfg.SigMirror = "cdn"
	if SignatureSource(cfg) != "official-cdn" {
		t.Fatal("cdn sentinel → official-cdn")
	}
}

func TestStaleness(t *testing.T) {
	f := &Feed{}
	f.setLastSuccess(time.Unix(1000, 0))
	if !f.Stale(time.Unix(1000+7200, 0), time.Hour) {
		t.Fatal("2h since success with 1h max → stale")
	}
	if f.Stale(time.Unix(1000+1800, 0), time.Hour) {
		t.Fatal("30m since success with 1h max → fresh")
	}
	if f.LastSuccess() != time.Unix(1000, 0) {
		t.Fatal("LastSuccess accessor")
	}
	var zero Feed
	if zero.Stale(time.Unix(1000, 0), time.Hour) {
		t.Fatal("never-succeeded feed is not reported stale (no baseline yet)")
	}
}

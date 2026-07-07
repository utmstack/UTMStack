package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

func TestScanFileMaliciousEmitsBrandedEvent(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "a.exe")
	if err := os.WriteFile(target, []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	spoolPath := filepath.Join(dir, "events.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	fq := &fakeQ{}
	s := New(config.Default(), c, sp, fq)
	s.scanBytes = func(addr string, data []byte) (bool, string, error) {
		return false, "Win.Test.EICAR_HDB-1", nil // pretend malicious
	}

	verdict, sig, err := s.ScanFile(target, event.SourceEngine)
	if err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	if verdict != cache.VerdictMalicious || !strings.Contains(sig, "EICAR") {
		t.Fatalf("verdict=%q sig=%q", verdict, sig)
	}
	// cache now holds the malicious verdict
	if rec, found, _ := c.Lookup(mustHash(t, target)); !found || rec.Verdict != cache.VerdictMalicious {
		t.Fatalf("verdict not cached: %+v found=%v", rec, found)
	}
	// spool holds one branded, clamav-free line
	b, _ := os.ReadFile(spoolPath)
	line := strings.TrimSpace(string(b))
	if !strings.Contains(line, `"product":"UTMStack EDR"`) {
		t.Fatalf("spool line not branded: %s", line)
	}
	if strings.Contains(strings.ToLower(line), "clam") {
		t.Fatalf("spool line leaked engine name: %s", line)
	}
	if !fq.called {
		t.Fatal("quarantiner was not invoked for malicious file")
	}
	if !strings.Contains(line, `"action":"quarantined"`) {
		t.Fatalf("expected quarantined action: %s", line)
	}
}

type fakeQ struct{ called bool }

func (f *fakeQ) Quarantine(path, sha256, detection string) (string, error) {
	f.called = true
	return "qid", nil
}

func mustHash(t *testing.T, p string) string {
	t.Helper()
	h, err := SHA256File(p)
	if err != nil {
		t.Fatal(err)
	}
	return h
}

// The scanner must stamp the signature-DB version a verdict was produced under,
// so the feed only re-scans a cached entry when signatures actually change.
func TestScanFileStampsSigDBVersion(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "clean.bin")
	if err := os.WriteFile(target, []byte("harmless"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := cache.Open(filepath.Join(dir, "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	sp, err := event.OpenSpool(filepath.Join(dir, "events.ndjson"), 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()

	s := New(config.Default(), c, sp, &fakeQ{})
	s.scanBytes = func(addr string, data []byte) (bool, string, error) { return true, "", nil } // clean
	s.sigDBVersion = func() string { return "27000" }

	if _, _, err := s.ScanFile(target, event.SourceEngine); err != nil {
		t.Fatalf("ScanFile: %v", err)
	}
	rec, found, _ := c.Lookup(mustHash(t, target))
	if !found || rec.SigDBVersion != "27000" {
		t.Fatalf("record not stamped with sig-db version: %+v (found=%v)", rec, found)
	}

	// A record stamped with the *current* version must NOT be re-invalidated.
	if n, _ := c.MarkStaleBySigDB("27000"); n != 0 {
		t.Fatalf("stamped record invalidated at same version (n=%d) — over-invalidation not fixed", n)
	}
	// It IS invalidated when the version actually changes.
	if n, _ := c.MarkStaleBySigDB("28000"); n != 1 {
		t.Fatalf("stamped record not invalidated on version change (n=%d)", n)
	}
}

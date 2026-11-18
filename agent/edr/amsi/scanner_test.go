package amsi

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func newTestScanner(t *testing.T, cfg config.EDRConfig) (*Scanner, string) {
	dir := t.TempDir()
	spoolPath := filepath.Join(dir, "e.ndjson")
	sp, err := event.OpenSpool(spoolPath, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = sp.Close() })
	s := NewScanner(cfg, sp, engine.New(cfg))
	return s, spoolPath
}

func TestNativeEngineScansAMSIUTF16Buffer(t *testing.T) {
	dir := t.TempDir()
	// Preserve raw buffer semantics: this signature targets UTF-16LE content
	// directly, matching what PowerShell submits through AMSI.
	if err := os.WriteFile(filepath.Join(dir, "script.ndb"), []byte("UTM.Test.Script:0:*:4d00410052004b0045005200\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir = dir
	eng := engine.New(cfg)
	// The service loads the shared engine before AMSI starts; scans never load it.
	if err := eng.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	s := NewScanner(cfg, nil, eng)
	if !s.ScanBuffer("PowerShell", []byte{'M', 0, 'A', 0, 'R', 0, 'K', 0, 'E', 0, 'R', 0}) {
		t.Fatal("native UTF-16LE buffer was not blocked")
	}
	if s.ScanBuffer("PowerShell", []byte("a benign script")) {
		t.Fatal("benign script blocked")
	}
}

func TestScanBufferMaliciousBlocksAndEmits(t *testing.T) {
	s, spoolPath := newTestScanner(t, config.Default())
	s.scanSnapshot = snapshotForTest(func(addr context.Context, data []byte) (bool, string, error) { return false, "PUA.Script.Test", nil })

	if block := s.ScanBuffer("PowerShell", []byte("bad")); !block {
		t.Fatal("expected block=true for malicious content")
	}
	b, _ := os.ReadFile(spoolPath)
	line := strings.TrimSpace(string(b))
	if !strings.Contains(line, `"source":"amsi"`) || !strings.Contains(line, `"action":"blocked"`) {
		t.Fatalf("bad amsi event: %s", line)
	}
	if strings.Contains(strings.ToLower(line), "clam") {
		t.Fatalf("leaked engine name: %s", line)
	}
}

func TestScanBufferCleanAllows(t *testing.T) {
	s, _ := newTestScanner(t, config.Default())
	s.scanSnapshot = snapshotForTest(func(addr context.Context, data []byte) (bool, string, error) { return true, "", nil })
	if block := s.ScanBuffer("PowerShell", []byte("ok")); block {
		t.Fatal("expected block=false for clean content")
	}
}

func TestScanBufferFailOpenVsClosed(t *testing.T) {
	open := config.Default() // FailMode=open
	s, _ := newTestScanner(t, open)
	s.scanSnapshot = snapshotForTest(func(addr context.Context, data []byte) (bool, string, error) {
		return false, "", errors.New("engine down")
	})
	if block := s.ScanBuffer("PowerShell", []byte("x")); block {
		t.Fatal("fail-open must allow on engine error")
	}

	closed := config.Default()
	closed.FailMode = "closed"
	s2, _ := newTestScanner(t, closed)
	s2.scanSnapshot = snapshotForTest(func(addr context.Context, data []byte) (bool, string, error) {
		return false, "", errors.New("engine down")
	})
	if block := s2.ScanBuffer("PowerShell", []byte("x")); !block {
		t.Fatal("fail-closed must block on engine error")
	}
}

// spoolLines returns the non-empty spool records written so far.
func spoolLines(t *testing.T, path string) []string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	var out []string
	for _, l := range strings.Split(string(b), "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

// Repeated incomplete analysis of the same buffer under the same generation is
// reported once per window, like the file scanner's noteIncomplete, instead of
// one health event (and one error log line) per buffer. The fail policy still
// applies to every buffer.
func TestIncompleteAnalysisIsDeduplicated(t *testing.T) {
	for name, dbFiles := range map[string]map[string]string{
		// Fresh endpoint: no generation provisioned yet.
		"engine-unavailable": {},
		// Loaded engine, content analysis incomplete.
		"content-incomplete": {"local.ndb": "UTM.Test:0:*:55544d2d544553542d4d41524b4552\n"},
	} {
		t.Run(name, func(t *testing.T) {
			dbDir := t.TempDir()
			for n, body := range dbFiles {
				if err := os.WriteFile(filepath.Join(dbDir, n), []byte(body), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			cfg := config.Default()
			cfg.SignatureDBDir = dbDir
			cfg.FailMode = "closed"
			eng := engine.New(cfg)
			loadErr := eng.EnsureRunning()
			if (len(dbFiles) == 0) != (loadErr != nil) {
				t.Fatalf("unexpected load result: %v", loadErr)
			}
			spoolPath := filepath.Join(t.TempDir(), "events.ndjson")
			sp, err := event.OpenSpool(spoolPath, 8<<20)
			if err != nil {
				t.Fatal(err)
			}
			defer sp.Close()
			s := NewScanner(cfg, sp, eng)
			// Recognized but unsupported content (a Symbian OS 9 package
			// header) is a genuine incomplete outcome independent of any
			// text-decoder behavior.
			script := []byte("\x7a\x1a\x20\x10 submitted buffer")
			for i := 0; i < 50; i++ {
				if !s.ScanBuffer("PowerShell", script) {
					t.Fatalf("buffer %d: fail-closed policy not applied", i)
				}
			}
			lines := spoolLines(t, spoolPath)
			if len(lines) != 1 {
				t.Fatalf("%d health events for 50 identical AMSI buffers (not deduplicated)", len(lines))
			}
			want := "Analysis.Incomplete.Unsupported"
			if len(dbFiles) == 0 {
				want = "Analysis.Incomplete.Unavailable"
			}
			if !strings.Contains(lines[0], `"action":"health"`) || !strings.Contains(lines[0], `"verdict":"unknown"`) ||
				!strings.Contains(lines[0], want) {
				t.Fatalf("health event lacks reason %s: %s", want, lines[0])
			}
		})
	}
}

// Deduplication is per content for content-dependent reasons, but a down
// engine is one condition: distinct buffers share a single Unavailable report.
func TestIncompleteReportIdentity(t *testing.T) {
	cases := map[string]struct {
		err  error
		want int
	}{
		"unsupported": {engine.ErrUnsupported, 3},
		"unavailable": {engine.ErrUnavailable, 1},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, spoolPath := newTestScanner(t, config.Default())
			s.scanSnapshot = snapshotForTest(func(context.Context, []byte) (bool, string, error) { return false, "", tc.err })
			for i := 0; i < 3; i++ {
				for j := 0; j < 2; j++ {
					_ = s.ScanBuffer("PowerShell", []byte{byte('a' + i)})
				}
			}
			if n := len(spoolLines(t, spoolPath)); n != tc.want {
				t.Fatalf("%d reports, want %d", n, tc.want)
			}
		})
	}
}

func TestIncompleteReportMemoryIsBounded(t *testing.T) {
	s, _ := newTestScanner(t, config.Default())
	s.scanSnapshot = snapshotForTest(func(context.Context, []byte) (bool, string, error) { return false, "", engine.ErrMalformed })
	for i := 0; i < maxIncompleteReports+10; i++ {
		_ = s.ScanBuffer("PowerShell", []byte(fmt.Sprint(i)))
	}
	if len(s.reported) != maxIncompleteReports {
		t.Fatalf("unbounded report deduplication: %d", len(s.reported))
	}
}

// While no generation is loaded, each AMSI buffer must fail fast instead of
// re-reading and recompiling the signature directory; Engine.Supervise owns
// retries.
func TestScanBufferDoesNotReloadSharedEngine(t *testing.T) {
	db := t.TempDir()
	if err := os.WriteFile(filepath.Join(db, "bad.ndb"), []byte("Bad.Rule:0:*:zz\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Default()
	cfg.SignatureDBDir = db
	eng := engine.New(cfg)
	if err := eng.EnsureRunning(); err == nil {
		t.Fatal("expected initial load failure")
	}
	s := NewScanner(cfg, nil, eng)
	// The loose database is fixed on disk; nobody calls Reload.
	if err := os.WriteFile(filepath.Join(db, "bad.ndb"), []byte("UTM.Test:0:*:4d41524b4552\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if s.ScanBuffer("PowerShell", []byte("MARKER")) || eng.Version() != "" {
			t.Fatalf("buffer %d reloaded the signature directory (version=%q)", i, eng.Version())
		}
	}
}

func TestIncompleteReasonNamesAnalysisNotPerformed(t *testing.T) {
	cause := fmt.Errorf("%w: %w", signatures.ErrRuleIncomplete, &signatures.UnknownError{Reason: signatures.UnknownPCRE})
	if got := incompleteReason(cause); got != "Unsupported.pcre" {
		t.Fatalf("reason = %q", got)
	}
}

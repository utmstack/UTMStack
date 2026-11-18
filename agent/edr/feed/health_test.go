package feed

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

func TestSignatureSource(t *testing.T) {
	cfg := config.Default()
	if SignatureSource(cfg) != "unconfigured" {
		t.Fatal("no server registered → unconfigured")
	}
	cfg.Server = "utm.example.com"
	if SignatureSource(cfg) != "utmstack-mirror" {
		t.Fatal("auto-derived mirror → utmstack-mirror")
	}
	cfg.SigMirror = "cdn"
	if SignatureSource(cfg) != "unconfigured" {
		t.Fatal("legacy cdn sentinel → unconfigured")
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

func TestUnavailableUpdaterReportsFailure(t *testing.T) {
	f := New(config.Default(), nil)
	if _, err := f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("missing engine accepted")
	}
	if f.LastError() == "" {
		t.Fatal("update failure not exposed")
	}
}

func TestClassifyUpdate(t *testing.T) {
	now := time.Unix(1_000_000, 0)
	interval := 4 * time.Hour
	recent, old := now.Add(-time.Hour), now.Add(-13*time.Hour)
	for _, tc := range []struct {
		name    string
		last    time.Time
		lastErr string
		active  bool
		want    string
	}{
		{"no attempt finished yet", time.Time{}, "", false, UpdatePending},
		{"refreshed recently", recent, "", true, UpdateCurrent},
		{"refresh failed, verified generation active", recent, "timeout", true, UpdateRefreshFailed},
		{"refresh failed, generation active, baseline unknown", time.Time{}, "timeout", true, UpdateRefreshFailed},
		{"refresh failing beyond three intervals", old, "timeout", true, UpdateStale},
		{"no refresh for three intervals", old, "", true, UpdateStale},
		{"no verified generation and refresh failed", time.Time{}, "timeout", false, UpdateUnavailable},
		{"generation gone and refresh failed", recent, "timeout", false, UpdateUnavailable},
	} {
		if got := classifyUpdate(now, tc.last, tc.lastErr, tc.active, interval); got != tc.want {
			t.Errorf("%s: got %q, want %q", tc.name, got, tc.want)
		}
	}
}

// deadMirror returns an HTTPS mirror URL on a port nothing listens on.
func deadMirror(t *testing.T) string {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := l.Addr().String()
	l.Close()
	return "https://" + addr + "/signatures"
}

func openSpool(t *testing.T) (*event.Spool, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "events.ndjson")
	sp, err := event.OpenSpool(path, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { sp.Close() })
	return sp, path
}

func spoolText(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		t.Fatal(err)
	}
	return string(b)
}

// The acceptance run restarted the service while the mirror was unreachable:
// a verified generation kept serving, yet status.json showed no last update
// and every start wrote a signature_updates_unavailable event.
func TestRestartWithVerifiedGenerationAndFailedRefresh(t *testing.T) {
	for _, persisted := range []bool{false, true} {
		t.Run(map[bool]string{false: "activation time", true: "persisted refresh"}[persisted], func(t *testing.T) {
			fix := newFeedFixture(t)
			cfg := fix.f.cfg
			cfg.SignatureDBDir = t.TempDir()
			fix.f.engine = engine.New(cfg)
			var c *cache.Cache
			if persisted {
				var err error
				if c, err = cache.Open(filepath.Join(t.TempDir(), "edr.db")); err != nil {
					t.Fatal(err)
				}
				defer c.Close()
				fix.f.cache = c
			}
			refreshed := time.Now().Add(-time.Hour).Truncate(time.Second)
			fix.f.now = func() time.Time { return refreshed }
			if _, err := fix.f.updateOnce(); err != nil {
				t.Fatal(err)
			}
			fix.f.noteHealth(4 * time.Hour)
			// The generation was activated before the last (no-op) refresh.
			activated := refreshed.Add(-time.Hour)
			if err := os.Chtimes(filepath.Join(cfg.SignatureDBDir, "CURRENT"), activated, activated); err != nil {
				t.Fatal(err)
			}

			// Restart: a new engine loads the stored generation; the mirror is down.
			restarted := engine.New(cfg)
			if err := restarted.Reload(); err != nil {
				t.Fatal(err)
			}
			down := cfg
			down.SigMirror = deadMirror(t)
			f := New(down, c, restarted)
			sp, spoolPath := openSpool(t)
			f.SetSpool(sp)
			f.RestoreHealth(restarted.DatabaseDir())
			if _, err := f.UpdateOnce(context.Background()); err == nil {
				t.Fatal("refresh from a closed port succeeded")
			}
			f.noteHealth(4 * time.Hour)

			want := activated
			if persisted {
				want = refreshed
			}
			if got := f.LastSuccess(); !got.Equal(want) {
				t.Fatalf("last update after restart = %v, want %v", got, want)
			}
			if got := f.UpdateState(); got != UpdateRefreshFailed {
				t.Fatalf("update state = %q, want %q", got, UpdateRefreshFailed)
			}
			if got := spoolText(t, spoolPath); got != "" {
				t.Fatalf("health event written for a failed refresh with a verified generation active:\n%s", got)
			}
			if got := f.LastError(); !strings.Contains(got, "connection refused") {
				t.Fatalf("refresh error %q does not name the cause", got)
			}
		})
	}
}

func TestGenerationActivatedAtMatchesOnlyTheActiveGeneration(t *testing.T) {
	fix := newFeedFixture(t)
	cfg := fix.f.cfg
	cfg.SignatureDBDir = t.TempDir()
	if got := GenerationActivatedAt(cfg.SignatureDBDir, 0, ""); !got.IsZero() {
		t.Fatalf("no generation reported activated at %v", got)
	}
	e := engine.New(cfg)
	fix.f.engine = e
	if _, err := fix.f.updateOnce(); err != nil {
		t.Fatal(err)
	}
	sequence, hash := e.FeedState()
	current, err := os.Stat(filepath.Join(cfg.SignatureDBDir, "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	if got := GenerationActivatedAt(cfg.SignatureDBDir, sequence, hash); !got.Equal(current.ModTime()) {
		t.Fatalf("activation time %v, want %v", got, current.ModTime())
	}
	// A pointer to a generation other than the one serving is not reported.
	if got := GenerationActivatedAt(cfg.SignatureDBDir, sequence+1, hash); !got.IsZero() {
		t.Fatalf("mismatched generation reported activated at %v", got)
	}
}

func TestFailedRefreshWithoutGenerationReportsUnavailableOnce(t *testing.T) {
	cfg := config.Default()
	cfg.SigMirror = deadMirror(t)
	cfg.SignaturePublicKey = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="
	f := New(cfg, nil, &testGenerationStore{})
	sp, spoolPath := openSpool(t)
	f.SetSpool(sp)
	f.RestoreHealth(t.TempDir())
	for i := 0; i < 2; i++ {
		if _, err := f.UpdateOnce(context.Background()); err == nil {
			t.Fatal("refresh from a closed port succeeded")
		}
		f.noteHealth(4 * time.Hour)
	}
	if got := f.UpdateState(); got != UpdateUnavailable {
		t.Fatalf("update state = %q, want %q", got, UpdateUnavailable)
	}
	if got := strings.Count(spoolText(t, spoolPath), "signature_updates_unavailable"); got != 1 {
		t.Fatalf("%d unavailable events, want 1", got)
	}
}

// A refresh that keeps failing for three intervals after the last success
// recorded before the restart is reported stale once, and a later success
// re-arms the warning.
func TestPersistedStaleBaselineWarnsOnceAndSuccessRearms(t *testing.T) {
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.SaveFeedState(cache.FeedState{Feed: signatureFeedState, LastFullSync: time.Now().Add(-48 * time.Hour)}); err != nil {
		t.Fatal(err)
	}
	fix := newFeedFixture(t)
	fix.store.sequence, fix.store.hash = 1, strings.Repeat("a", 64)
	up := fix.f
	cfg := up.cfg
	cfg.SigMirror = deadMirror(t)
	down := New(cfg, c, fix.store)
	sp, spoolPath := openSpool(t)
	down.SetSpool(sp)
	down.RestoreHealth(t.TempDir())
	for i := 0; i < 2; i++ {
		_, _ = down.UpdateOnce(context.Background())
		down.noteHealth(4 * time.Hour)
	}
	if got := down.UpdateState(); got != UpdateStale {
		t.Fatalf("update state = %q, want %q", got, UpdateStale)
	}
	events := spoolText(t, spoolPath)
	if strings.Count(events, "signature_updates_stale") != 1 || strings.Contains(events, "signature_updates_unavailable") {
		t.Fatalf("want exactly one stale event:\n%s", events)
	}
	// The mirror is back: the next cycle succeeds and re-arms the warning.
	down.client, down.cfg = up.client, up.cfg
	if _, err := down.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}
	down.noteHealth(4 * time.Hour)
	if got := down.UpdateState(); got != UpdateCurrent {
		t.Fatalf("update state after success = %q, want %q", got, UpdateCurrent)
	}
	if st, err := c.GetFeedState(signatureFeedState); err != nil || time.Since(st.LastFullSync) > time.Minute {
		t.Fatalf("successful refresh not persisted: %+v %v", st, err)
	}
}

func TestRequestFailureCause(t *testing.T) {
	dial := func(err error) error {
		return &url.Error{Op: "Get", URL: "https://mirror/manifest.json", Err: &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connectex", err)}}
	}
	tlsErr := func(err error) error {
		return &url.Error{Op: "Get", URL: "https://mirror/manifest.json", Err: &tls.CertificateVerificationError{Err: err}}
	}
	for _, tc := range []struct {
		err  error
		want string
	}{
		{tlsErr(x509.UnknownAuthorityError{}), "untrusted certificate"},
		{tlsErr(x509.HostnameError{Host: "mirror"}), "certificate name mismatch"},
		{tlsErr(x509.CertificateInvalidError{Reason: x509.Expired}), "certificate expired or not yet valid"},
		{tlsErr(x509.CertificateInvalidError{Reason: x509.IncompatibleUsage}), "certificate not valid for this server"},
		{tlsErr(errors.New("x509: certificate verification failed")), "certificate verification failed"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: http.ErrSchemeMismatch}, "server does not use TLS"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: tls.RecordHeaderError{Msg: "first record does not look like a TLS handshake"}}, "server does not use TLS"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: &net.OpError{Op: "remote error", Err: errors.New("tls: handshake failure")}}, "TLS handshake rejected"},
		{dial(syscall.ECONNREFUSED), "connection refused"},
		{dial(syscall.Errno(10061)), "connection refused"}, // WSAECONNREFUSED
		{dial(syscall.Errno(10060)), "timeout"},            // WSAETIMEDOUT
		{dial(syscall.Errno(10054)), "connection reset"},   // WSAECONNRESET
		{dial(syscall.Errno(10065)), "network unreachable"},
		{dial(syscall.Errno(10051)), "network unreachable"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: &net.DNSError{Err: "no such host", Name: "mirror", IsNotFound: true}}, "name resolution failed"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: context.DeadlineExceeded}, "timeout"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: context.Canceled}, "canceled"},
		{&url.Error{Op: "Get", URL: "https://mirror", Err: io.ErrUnexpectedEOF}, "connection closed"},
		{errors.New("something else"), "network error"},
	} {
		got := requestFailureCause(tc.err)
		if got != tc.want {
			t.Errorf("requestFailureCause(%v) = %q, want %q", tc.err, got, tc.want)
		}
		if strings.Contains(strings.ToLower(got), "clam") {
			t.Errorf("cause %q is not branding-safe", got)
		}
	}
}

// A mirror whose certificate the host does not trust is reported as such, not
// only as a failed request.
func TestUntrustedMirrorCertificateIsNamed(t *testing.T) {
	fix := newFeedFixture(t)
	f := NewNative(fix.f.cfg, nil, fix.store) // production client: fixture CA not trusted
	if _, err := f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("untrusted mirror accepted")
	}
	if got := f.LastError(); !strings.Contains(got, "untrusted certificate") {
		t.Fatalf("error %q does not name the certificate problem", got)
	}
}

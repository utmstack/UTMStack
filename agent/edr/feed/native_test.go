package feed

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

type testGenerationStore struct {
	sequence uint64
	hash     string
	version  string
	files    map[string][]byte
	calls    int
	err      error
}

func (s *testGenerationStore) FeedState() (uint64, string) { return s.sequence, s.hash }
func (s *testGenerationStore) Version() string             { return s.version }
func (s *testGenerationStore) ActivateGeneration(files map[string][]byte, sequence uint64, hash string) (string, error) {
	s.calls++
	if s.err != nil {
		return "", s.err
	}
	s.files, s.sequence, s.hash, s.version = files, sequence, hash, "native-test-200"
	return s.version, nil
}

type feedFixture struct {
	f         *Feed
	store     *testGenerationStore
	manifest  Manifest
	private   ed25519.PrivateKey
	resources map[string][]byte
	// handlers replace the static resource for a path, for transfer timing tests.
	handlers map[string]http.HandlerFunc
	// artifactHits counts requests for anything other than the manifest.
	artifactHits atomic.Int64
}

func newFeedFixture(t *testing.T) *feedFixture {
	t.Helper()
	public, private, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	data := []byte("Test.Rule:0:*:41424344\n")
	hash := sha256.Sum256(data)
	fix := &feedFixture{
		store:     &testGenerationStore{version: "previous"},
		private:   private,
		manifest:  Manifest{Version: "release-2", Sequence: 2, ExpiresAt: time.Now().Add(time.Hour), Files: []ManifestFile{{Name: "test.ndb", URL: "test.ndb", SHA256: hex.EncodeToString(hash[:])}}},
		resources: map[string][]byte{"/signatures/test.ndb": data},
	}
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/signatures/manifest.json" {
			fix.artifactHits.Add(1)
		}
		if h := fix.handlers[r.URL.Path]; h != nil {
			h(w, r)
			return
		}
		data, exists := fix.resources[r.URL.Path]
		if !exists {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(data)
	}))
	t.Cleanup(server.Close)
	cfg := config.Default()
	cfg.SigMirror = server.URL + "/signatures"
	cfg.SignaturePublicKey = base64.StdEncoding.EncodeToString(public)
	fix.f = NewNative(cfg, nil, fix.store)
	// Trust the fixture certificate but keep the production client limits.
	client := server.Client()
	client.CheckRedirect = fix.f.client.CheckRedirect
	client.Timeout = fix.f.client.Timeout
	fix.f.client = client
	fix.sign(t)
	return fix
}

func (f *feedFixture) sign(t *testing.T) {
	t.Helper()
	payload, err := json.Marshal(f.manifest)
	if err != nil {
		t.Fatal(err)
	}
	envelope := ManifestEnvelope{Payload: base64.StdEncoding.EncodeToString(payload), Signature: base64.StdEncoding.EncodeToString(ed25519.Sign(f.private, payload))}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatal(err)
	}
	f.resources["/signatures/manifest.json"] = encoded
}

func TestSignedUpdateMarksCachedVerdictsStale(t *testing.T) {
	fix := newFeedFixture(t)
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	if err := c.Store(cache.VerdictRecord{SHA256: "clean1", Verdict: cache.VerdictClean, SigDBVersion: "previous"}); err != nil {
		t.Fatal(err)
	}
	fix.f.cache = c
	version, err := fix.f.updateOnce()
	if err != nil || version != fix.store.version {
		t.Fatalf("version=%q error=%v", version, err)
	}
	if record, _, _ := c.Lookup("clean1"); !record.Stale {
		t.Fatal("old clean verdict was not invalidated")
	}
	if fix.store.sequence != 2 || fix.store.hash == "" || fix.f.LastSuccess().IsZero() {
		t.Fatal("activation or update health missing")
	}
	if _, err := fix.f.updateOnce(); err != nil {
		t.Fatalf("identical signed sequence must be retryable: %v", err)
	}
}

// A cycle whose signed manifest is the active generation has nothing to
// download, compile or rewrite, but it still shows the mirror is current and
// still retries clean-verdict invalidation for the active generation.
func TestUnchangedManifestIsANoOpUpdate(t *testing.T) {
	fix := newFeedFixture(t)
	c, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()
	fix.f.cache = c
	clock := time.Now()
	fix.f.now = func() time.Time { return clock }
	first, err := fix.f.updateOnce()
	if err != nil {
		t.Fatal(err)
	}
	// A clean verdict from another generation, as if invalidation had failed
	// after the previous activation.
	if err := c.Store(cache.VerdictRecord{SHA256: "clean2", Verdict: cache.VerdictClean, SigDBVersion: "previous"}); err != nil {
		t.Fatal(err)
	}
	clock = clock.Add(30 * time.Minute)
	second, err := fix.f.updateOnce()
	if err != nil || second != first {
		t.Fatalf("identical manifest must be a successful no-op: version %q -> %q, error %v", first, second, err)
	}
	if hits := fix.artifactHits.Load(); hits != 1 || fix.store.calls != 1 {
		t.Fatalf("unchanged generation re-downloaded or re-activated: %d artifact requests, %d activations", hits, fix.store.calls)
	}
	if !fix.f.LastSuccess().Equal(clock) {
		t.Fatal("no-op cycle did not refresh update health")
	}
	if record, _, _ := c.Lookup("clean2"); !record.Stale {
		t.Fatal("no-op cycle skipped clean-verdict invalidation")
	}
	fix.manifest.Sequence, fix.manifest.Version = 3, "release-3"
	fix.sign(t)
	if _, err := fix.f.updateOnce(); err != nil || fix.artifactHits.Load() != 2 || fix.store.calls != 2 || fix.store.sequence != 3 {
		t.Fatalf("newer sequence was not downloaded and activated: %v", err)
	}
}

// Go's http.Client.Timeout also covers reading the response body, so a
// whole-request limit aborts a large artifact on a slow link while bytes are
// still arriving. A progressing transfer is bounded by the update budget only.
func TestSlowButProgressingArtifactIsNotAborted(t *testing.T) {
	fix := newFeedFixture(t)
	if fix.f.client.Timeout != 0 {
		t.Fatalf("whole-request client timeout %s also limits artifact body reads", fix.f.client.Timeout)
	}
	fix.f.stallTimeout = 400 * time.Millisecond
	data := fix.resources["/signatures/test.ndb"]
	fix.handlers = map[string]http.HandlerFunc{"/signatures/test.ndb": func(w http.ResponseWriter, r *http.Request) {
		// Every gap is shorter than the stall limit; the whole transfer is
		// several times longer than it.
		for off := 0; off < len(data); off += 2 {
			if _, err := w.Write(data[off:min(off+2, len(data))]); err != nil {
				return
			}
			w.(http.Flusher).Flush()
			select {
			case <-r.Context().Done():
				return
			case <-time.After(100 * time.Millisecond):
			}
		}
	}}
	start := time.Now()
	if _, err := fix.f.updateOnce(); err != nil {
		t.Fatalf("progressing download aborted after %s: %v", time.Since(start).Round(time.Millisecond), err)
	}
	if elapsed := time.Since(start); elapsed < 2*fix.f.stallTimeout || fix.store.calls != 1 {
		t.Fatalf("transfer took %s and activated %d times; the fixture did not exercise a long transfer", elapsed, fix.store.calls)
	}
}

// Receiving the response headers is progress: the wait for the first body
// bytes is its own gap, not added to the wait for the headers.
func TestResponseHeadersReArmTheStallLimit(t *testing.T) {
	fix := newFeedFixture(t)
	fix.f.stallTimeout = 600 * time.Millisecond
	data := fix.resources["/signatures/test.ndb"]
	pause := func(r *http.Request) bool {
		select {
		case <-r.Context().Done():
			return false
		case <-time.After(400 * time.Millisecond):
			return true
		}
	}
	fix.handlers = map[string]http.HandlerFunc{"/signatures/test.ndb": func(w http.ResponseWriter, r *http.Request) {
		// Each gap is shorter than the stall limit; together they exceed it.
		if !pause(r) {
			return
		}
		w.WriteHeader(http.StatusOK)
		w.(http.Flusher).Flush()
		if !pause(r) {
			return
		}
		_, _ = w.Write(data)
	}}
	start := time.Now()
	if _, err := fix.f.updateOnce(); err != nil {
		t.Fatalf("transfer with a late but prompt body aborted after %s: %v", time.Since(start).Round(time.Millisecond), err)
	}
	if fix.store.calls != 1 {
		t.Fatalf("generation activated %d times", fix.store.calls)
	}
}

// A mirror that stops sending, before the headers or in the middle of a body,
// fails the update after the stall limit instead of holding it for the whole
// update budget.
func TestStalledTransferAbortsWithoutActivation(t *testing.T) {
	block := func(r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-time.After(10 * time.Second):
		}
	}
	for _, tc := range []struct {
		name    string
		path    string
		handler func([]byte) http.HandlerFunc
	}{
		{"manifest headers", "/signatures/manifest.json", func([]byte) http.HandlerFunc {
			return func(_ http.ResponseWriter, r *http.Request) { block(r) }
		}},
		{"artifact headers", "/signatures/test.ndb", func([]byte) http.HandlerFunc {
			return func(_ http.ResponseWriter, r *http.Request) { block(r) }
		}},
		{"artifact body", "/signatures/test.ndb", func(data []byte) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write(data[:len(data)/2])
				w.(http.Flusher).Flush()
				block(r)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fix := newFeedFixture(t)
			fix.f.stallTimeout = 300 * time.Millisecond
			fix.handlers = map[string]http.HandlerFunc{tc.path: tc.handler(fix.resources[tc.path])}
			start := time.Now()
			_, err := fix.f.updateOnce()
			if err == nil || !strings.Contains(err.Error(), "stalled") {
				t.Fatalf("stalled transfer not reported as stalled: %v", err)
			}
			if elapsed := time.Since(start); elapsed > 5*time.Second {
				t.Fatalf("stall detected only after %s", elapsed)
			}
			if fix.store.calls != 0 || !fix.f.LastSuccess().IsZero() {
				t.Fatal("stalled update altered the active generation or health")
			}
		})
	}
}

func TestRejectedUpdatesPreserveActiveGeneration(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*feedFixture)
	}{
		{"missing key", func(f *feedFixture) { f.f.cfg.SignaturePublicKey = "" }},
		{"wrong key", func(f *feedFixture) {
			key, _, _ := ed25519.GenerateKey(rand.Reader)
			f.f.cfg.SignaturePublicKey = base64.StdEncoding.EncodeToString(key)
		}},
		{"expired", func(f *feedFixture) { f.manifest.ExpiresAt = time.Now().Add(-time.Minute) }},
		{"rollback", func(f *feedFixture) { f.store.sequence = 3 }},
		{"sequence reuse", func(f *feedFixture) { f.store.sequence = 2; f.store.hash = "different" }},
		{"zero sequence", func(f *feedFixture) { f.manifest.Sequence = 0 }},
		{"empty release", func(f *feedFixture) { f.manifest.Version = "" }},
		{"no files", func(f *feedFixture) { f.manifest.Files = nil }},
		{"duplicate files", func(f *feedFixture) { f.manifest.Files = append(f.manifest.Files, f.manifest.Files[0]) }},
		{"invalid checksum", func(f *feedFixture) { f.manifest.Files[0].SHA256 = "bad" }},
		{"checksum mismatch", func(f *feedFixture) { f.manifest.Files[0].SHA256 = strings.Repeat("0", 64) }},
		{"fetch failure", func(f *feedFixture) { delete(f.resources, "/signatures/test.ndb") }},
		{"traversal", func(f *feedFixture) { f.manifest.Files[0].URL = "../test.ndb" }},
		{"encoded traversal", func(f *feedFixture) { f.manifest.Files[0].URL = "%2e%2e/test.ndb" }},
		{"external url", func(f *feedFixture) { f.manifest.Files[0].URL = "https://example.com/rules" }},
		{"protocol relative", func(f *feedFixture) { f.manifest.Files[0].URL = "//example.com/rules" }},
		{"absolute path", func(f *feedFixture) { f.manifest.Files[0].URL = "/rules" }},
		{"invalid name", func(f *feedFixture) { f.manifest.Files[0].Name = "../rules.ndb" }},
		{"unknown format", func(f *feedFixture) { f.manifest.Files[0].Format = "invented" }},
		{"legacy cdn", func(f *feedFixture) { f.f.cfg.SigMirror = "cdn" }},
		{"insecure mirror", func(f *feedFixture) { f.f.cfg.SigMirror = "http://localhost/rules"; f.f.cfg.SkipCertValidate = true }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fix := newFeedFixture(t)
			tc.mutate(fix)
			fix.sign(t)
			if _, err := fix.f.updateOnce(); err == nil {
				t.Fatal("invalid update accepted")
			}
			if fix.store.calls != 0 || fix.store.version != "previous" || !fix.f.LastSuccess().IsZero() {
				t.Fatal("rejected update altered active generation or health")
			}
		})
	}
}

func TestInterruptedDownloadDoesNotActivatePartialGeneration(t *testing.T) {
	fix := newFeedFixture(t)
	fix.manifest.Files = append(fix.manifest.Files, ManifestFile{Name: "missing.ndb", URL: "missing.ndb", SHA256: strings.Repeat("0", 64)})
	fix.sign(t)
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("incomplete generation accepted")
	}
	if fix.store.calls != 0 {
		t.Fatal("partial generation activated")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := fix.f.UpdateOnce(ctx); err == nil {
		t.Fatal("canceled update accepted")
	}
	if fix.store.calls != 0 {
		t.Fatal("canceled update activated")
	}
}

func TestCompilationErrorRetainsGeneration(t *testing.T) {
	fix := newFeedFixture(t)
	fix.store.err = errors.New("unsupported rule")
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("compile error hidden")
	}
	if fix.store.version != "previous" || fix.store.sequence != 0 || !fix.f.LastSuccess().IsZero() {
		t.Fatal("failed compilation recorded success")
	}
}

func TestTamperedAndTrailingManifestRejected(t *testing.T) {
	fix := newFeedFixture(t)
	var envelope ManifestEnvelope
	_ = json.Unmarshal(fix.resources["/signatures/manifest.json"], &envelope)
	envelope.Payload = base64.StdEncoding.EncodeToString([]byte(`{"version":"tampered"}`))
	fix.resources["/signatures/manifest.json"], _ = json.Marshal(envelope)
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("tampered signed payload accepted")
	}
	fix.sign(t)
	fix.resources["/signatures/manifest.json"] = append(fix.resources["/signatures/manifest.json"], []byte(`{}`)...)
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("trailing envelope data accepted")
	}
}

func TestReadBoundedRejectsOversize(t *testing.T) {
	if _, err := readBounded(strings.NewReader("12345"), 4); err == nil {
		t.Fatal("oversize accepted")
	}
	if got, err := readBounded(strings.NewReader("1234"), 4); err != nil || !bytes.Equal(got, []byte("1234")) {
		t.Fatal("exact bound rejected")
	}
}

func TestNextDelayRetriesFasterAfterFailure(t *testing.T) {
	if d := nextDelay(nil, 4*time.Hour); d != 4*time.Hour {
		t.Fatal(d)
	}
	if d := nextDelay(errors.New("x"), 4*time.Hour); d != 15*time.Minute {
		t.Fatal(d)
	}
	if d := nextDelay(errors.New("x"), 5*time.Minute); d != 5*time.Minute {
		t.Fatal(d)
	}
}

func TestNativeMirrorUsesVerifiedHTTPS(t *testing.T) {
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.SkipCertValidate = true
	got, err := NativeMirrorURL(cfg)
	if err != nil || got != "https://utm.example.com:9001/private/edr/native" {
		t.Fatalf("%s %v", got, err)
	}
	f := New(cfg, nil)
	if f.client.Transport != nil {
		t.Fatal("custom insecure TLS transport installed")
	}
	if err := f.client.CheckRedirect(nil, nil); err == nil {
		t.Fatal("redirects must be rejected")
	}
}

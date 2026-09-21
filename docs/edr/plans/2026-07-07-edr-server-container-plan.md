# UTMStack EDR — Master-Server `edr` Container + Client Polish — Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Stand up the master-server `edr` container (ClamAV signature mirror via `cvdupdate` + ThreatWinds feed mirror, normalized into a shared volume served by `agentmanager` at `https://<server>:9001/private/edr/**`) and finish the agent-side client polish (auto-derived URLs, path-based checksums, signature failover, CA bundle) so both channels work out of the box.

**Architecture:** The `edr` container is a producer, not a server: a Go daemon runs `cvdupdate` on a schedule into `/mirror/signatures/` and pulls+normalizes ThreatWinds accumulative lists into `/mirror/feeds/v1/…` with `.sha256` siblings; `agentmanager` serves the volume read-only as static files. Agents consume via freshclam `PrivateMirror` and the existing `edr/netblock` feed client. Spec: `docs/superpowers/specs/2026-07-07-edr-server-container-design.md`.

**Tech Stack:** Go (new module `github.com/utmstack/UTMStack/edr`; agent module unchanged Go 1.25.5), `cvdupdate` (Python, official ClamAV mirror tool — the container's one non-Go component), gin (existing, agent-manager), Docker/Swarm via the existing installer.

## Global Constraints

- **The user drives git.** Do NOT `git commit`, do NOT create branches, NEVER revert. End every task with green tests and leave changes uncommitted. (Plan steps therefore end at "tests pass" — there are no commit steps.)
- Work in `/Users/atlas/UTMStack/EDR/utmstack-v12/` (branch `edr-phase1` checkout of `release/v12.0.0`). Agent-side Go commands run from `utmstack-v12/agent/`; server module from `utmstack-v12/edr/`; agent-manager from `utmstack-v12/agent-manager/`; installer from `utmstack-v12/installer/`.
- **Branding (agent side):** `clamav`/`clamd`/`threatwinds` must never appear in an emitted event field or surfaced agent log/status string. The new mirror URL paths deliberately avoid vendor names (`/private/edr/signatures`, `/private/edr/feeds/v1`). Server-side code/logs (the `edr` container, platform services) may name vendors freely.
- Agent module: build + vet clean on **windows/amd64 and arm64**; all `edr` unit tests green on darwin. Do **not** run `go mod tidy` on macOS (it drops the windows-only `github.com/0xrawsec/golang-etw`).
- TDD: failing test → minimal implementation → pass, per task.
- Labeling audit must stay green: `grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/` and the same for `threatwinds`, run in `utmstack-v12/agent/`.
- No new Go dependencies in the agent module. The server module may use stdlib only.

---

# Part A — server-side `edr/` module

### Task 1: Widen the sparse checkout + scaffold the `edr/` Go module

**Files:**
- Create: `utmstack-v12/edr/go.mod`, `utmstack-v12/edr/main.go`

**Interfaces:**
- Produces: an empty but buildable module `github.com/utmstack/UTMStack/edr`; working trees for `agent-manager/`, `installer/`, `.github/` (needed by Tasks 10–12).

- [ ] **Step 1: Widen the sparse checkout**

```bash
cd /Users/atlas/UTMStack/EDR/utmstack-v12
git sparse-checkout add agent-manager installer .github
ls agent-manager/updates/updates.go installer/docker/compose.go .github/workflows/v12-deployment-pipeline.yml
```
Expected: all three files exist. (This is a read-only git config operation — allowed; it fetches blobs from the blobless clone.)

- [ ] **Step 2: Scaffold the module**

`utmstack-v12/edr/go.mod`:
```
module github.com/utmstack/UTMStack/edr

go 1.25.5
```

`utmstack-v12/edr/main.go`:
```go
package main

func main() {
	// Wired up in the daemon task.
}
```

- [ ] **Step 3: Verify it builds**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go build ./... && go vet ./...`
Expected: no output, exit 0.

---

### Task 2: `mirror` — atomic artifact writer with `.sha256` sibling

**Files:**
- Create: `utmstack-v12/edr/internal/mirror/writer.go`
- Test: `utmstack-v12/edr/internal/mirror/writer_test.go`

**Interfaces:**
- Produces: `func WriteArtifact(root, rel string, data []byte) error` — atomically writes `root/rel` then `root/rel.sha256` (bare lowercase sha256-hex of `data`). Artifact renamed into place **before** the checksum so a reader can never observe a new checksum with an old artifact (worst case: mismatch → the agent skips and retries next cycle). Also `func atomicWrite(path string, data []byte) error` (package-private, reused by the recorder in Task 3).

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/mirror/writer_test.go
package mirror

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteArtifactWritesDataAndChecksumSibling(t *testing.T) {
	root := t.TempDir()
	data := []byte("1.2.3.4\n5.6.7.8\n")
	rel := "feeds/v1/download/list/level1/accumulative/ip"
	if err := WriteArtifact(root, rel, data); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil || string(got) != string(data) {
		t.Fatalf("artifact = %q, err %v", got, err)
	}
	sum, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)+".sha256"))
	if err != nil {
		t.Fatal(err)
	}
	want := sha256.Sum256(data)
	if string(sum) != hex.EncodeToString(want[:]) {
		t.Fatalf("checksum = %q, want %q", sum, hex.EncodeToString(want[:]))
	}
}

func TestWriteArtifactOverwritesAndLeavesNoTempFiles(t *testing.T) {
	root := t.TempDir()
	if err := WriteArtifact(root, "a/b", []byte("old")); err != nil {
		t.Fatal(err)
	}
	if err := WriteArtifact(root, "a/b", []byte("new")); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(filepath.Join(root, "a", "b"))
	if string(got) != "new" {
		t.Fatalf("artifact = %q, want new", got)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "a"))
	for _, e := range entries {
		if filepath.Ext(e.Name()) == ".tmp" {
			t.Fatalf("temp file left behind: %s", e.Name())
		}
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/mirror/ -v`
Expected: FAIL — `undefined: WriteArtifact`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/mirror/writer.go
// Package mirror maintains the static mirror tree the platform serves to
// agents at https://<server>:9001/private/edr/ — artifacts are written
// atomically and never removed on failure (stale, never empty).
package mirror

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
)

// WriteArtifact atomically writes data to root/rel and a sibling
// root/rel.sha256 holding the bare lowercase sha256 hex of data (the exact
// format the agent's feed client verifies). The artifact lands before the
// checksum, so a torn read can only produce a mismatch (agent skips and
// retries), never a false verification.
func WriteArtifact(root, rel string, data []byte) error {
	dst := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := atomicWrite(dst, data); err != nil {
		return err
	}
	sum := sha256.Sum256(data)
	return atomicWrite(dst+".sha256", []byte(hex.EncodeToString(sum[:])))
}

func atomicWrite(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/mirror/ -v`
Expected: PASS (2 tests).

---

### Task 3: `mirror` — status recorder (`status.json`)

**Files:**
- Create: `utmstack-v12/edr/internal/mirror/status.go`
- Test: `utmstack-v12/edr/internal/mirror/status_test.go`

**Interfaces:**
- Consumes: `atomicWrite` (Task 2).
- Produces:
  - `type SigStatus struct { LastSuccess, LastAttempt time.Time; LastError string; Databases map[string]string }` (json: `last_success`, `last_attempt`, `last_error`, `databases`)
  - `type FeedStatus struct { LastSuccess, LastAttempt time.Time; Indicators int; SHA256 string; LastError string }` (json: `last_success`, `last_attempt`, `indicators`, `sha256`, `last_error`)
  - `func NewRecorder(root string) *Recorder`
  - `func (r *Recorder) SigSuccess(dbs map[string]string)` / `SigFailure(errMsg string)`
  - `func (r *Recorder) FeedSuccess(key string, indicators int, sha string)` / `FeedFailure(key, errMsg string)` — failure updates `last_attempt`/`last_error` but **preserves** the previous success fields (stale, never empty).
  - Every mutation stamps `updated` and rewrites `root/status.json` atomically. Mutex-guarded (the signature and feed loops run concurrently).

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/mirror/status_test.go
package mirror

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func readStatus(t *testing.T, root string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func TestRecorderWritesFeedSuccessAndPreservesItOnFailure(t *testing.T) {
	root := t.TempDir()
	r := NewRecorder(root)
	r.FeedSuccess("level1/ip", 42, "abc123")
	m := readStatus(t, root)
	feeds := m["feeds"].(map[string]any)
	f := feeds["level1/ip"].(map[string]any)
	if f["indicators"].(float64) != 42 || f["sha256"] != "abc123" {
		t.Fatalf("feed status = %v", f)
	}
	lastSuccess := f["last_success"]

	r.FeedFailure("level1/ip", "upstream 503")
	m = readStatus(t, root)
	f = m["feeds"].(map[string]any)["level1/ip"].(map[string]any)
	if f["last_error"] != "upstream 503" {
		t.Fatalf("last_error = %v", f["last_error"])
	}
	if f["last_success"] != lastSuccess || f["indicators"].(float64) != 42 {
		t.Fatal("failure must preserve previous success fields")
	}
}

func TestRecorderWritesSignatureStatus(t *testing.T) {
	root := t.TempDir()
	r := NewRecorder(root)
	r.SigSuccess(map[string]string{"daily.cvd": "27461"})
	m := readStatus(t, root)
	sig := m["signatures"].(map[string]any)
	if sig["databases"].(map[string]any)["daily.cvd"] != "27461" {
		t.Fatalf("signatures = %v", sig)
	}
	r.SigFailure("cdn unreachable")
	m = readStatus(t, root)
	sig = m["signatures"].(map[string]any)
	if sig["last_error"] != "cdn unreachable" {
		t.Fatalf("last_error = %v", sig["last_error"])
	}
	if m["updated"] == nil {
		t.Fatal("updated must be stamped")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/mirror/ -run Recorder -v`
Expected: FAIL — `undefined: NewRecorder`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/mirror/status.go
package mirror

import (
	"encoding/json"
	"path/filepath"
	"sync"
	"time"
)

type SigStatus struct {
	LastSuccess time.Time         `json:"last_success"`
	LastAttempt time.Time         `json:"last_attempt"`
	LastError   string            `json:"last_error"`
	Databases   map[string]string `json:"databases,omitempty"`
}

type FeedStatus struct {
	LastSuccess time.Time `json:"last_success"`
	LastAttempt time.Time `json:"last_attempt"`
	Indicators  int       `json:"indicators"`
	SHA256      string    `json:"sha256"`
	LastError   string    `json:"last_error"`
}

type status struct {
	Signatures SigStatus             `json:"signatures"`
	Feeds      map[string]FeedStatus `json:"feeds"`
	Updated    time.Time             `json:"updated"`
}

// Recorder maintains status.json at the mirror root — the platform's single
// point to monitor mirror freshness. Safe for concurrent use by the signature
// and feed sync loops.
type Recorder struct {
	mu   sync.Mutex
	root string
	st   status
}

func NewRecorder(root string) *Recorder {
	return &Recorder{root: root, st: status{Feeds: map[string]FeedStatus{}}}
}

func (r *Recorder) SigSuccess(dbs map[string]string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	r.st.Signatures.LastSuccess = now
	r.st.Signatures.LastAttempt = now
	r.st.Signatures.LastError = ""
	r.st.Signatures.Databases = dbs
	r.write()
}

func (r *Recorder) SigFailure(errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.st.Signatures.LastAttempt = time.Now().UTC()
	r.st.Signatures.LastError = errMsg
	r.write()
}

func (r *Recorder) FeedSuccess(key string, indicators int, sha string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	r.st.Feeds[key] = FeedStatus{LastSuccess: now, LastAttempt: now, Indicators: indicators, SHA256: sha}
	r.write()
}

// FeedFailure records the attempt + error but preserves the previous success
// fields — the mirror content itself is likewise never removed on failure.
func (r *Recorder) FeedFailure(key, errMsg string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	f := r.st.Feeds[key]
	f.LastAttempt = time.Now().UTC()
	f.LastError = errMsg
	r.st.Feeds[key] = f
	r.write()
}

func (r *Recorder) write() {
	r.st.Updated = time.Now().UTC()
	b, err := json.MarshalIndent(r.st, "", "  ")
	if err != nil {
		return
	}
	_ = atomicWrite(filepath.Join(r.root, "status.json"), b)
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/mirror/ -v`
Expected: PASS (all mirror tests).

---

### Task 4: `feeds` — normalizer (upstream artifact → agent wire format)

**Files:**
- Create: `utmstack-v12/edr/internal/feeds/normalize.go`
- Test: `utmstack-v12/edr/internal/feeds/normalize_test.go`

**Interfaces:**
- Produces: `func Normalize(upstream []byte) (artifact []byte, count int, err error)` — accepts a gzip'd tarball (members concatenated) **or** an already-gzip'd plain list; returns a gzip'd, newline-separated, deduped, **sorted** list (sorted ⇒ deterministic bytes ⇒ stable sha256) and the indicator count. Blank lines and `#` comments dropped. Non-gzip input is an error.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/feeds/normalize_test.go
package feeds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"io"
	"testing"
)

func gzBytes(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	return b.Bytes()
}

func tarGzBytes(t *testing.T, members map[string]string) []byte {
	t.Helper()
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	tw := tar.NewWriter(gw)
	for name, content := range members {
		if err := tw.WriteHeader(&tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	_ = tw.Close()
	_ = gw.Close()
	return b.Bytes()
}

func gunzipLines(t *testing.T, artifact []byte) string {
	t.Helper()
	r, err := gzip.NewReader(bytes.NewReader(artifact))
	if err != nil {
		t.Fatal(err)
	}
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func TestNormalizeTarGzMergesDedupesSorts(t *testing.T) {
	in := tarGzBytes(t, map[string]string{
		"part1": "9.9.9.9\n1.1.1.1\n# comment\n",
		"part2": "1.1.1.1\n2.2.2.2\n\n",
	})
	artifact, n, err := Normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := gunzipLines(t, artifact); got != "1.1.1.1\n2.2.2.2\n9.9.9.9\n" {
		t.Fatalf("artifact lines = %q", got)
	}
	if n != 3 {
		t.Fatalf("count = %d, want 3", n)
	}
}

func TestNormalizePlainGzipListPassesThrough(t *testing.T) {
	in := gzBytes(t, "b.example.com\na.example.com\na.example.com\n")
	artifact, n, err := Normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got := gunzipLines(t, artifact); got != "a.example.com\nb.example.com\n" {
		t.Fatalf("artifact lines = %q", got)
	}
	if n != 2 {
		t.Fatalf("count = %d, want 2", n)
	}
}

func TestNormalizeRejectsNonGzip(t *testing.T) {
	if _, _, err := Normalize([]byte("not gzip")); err == nil {
		t.Fatal("expected error for non-gzip input")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/feeds/ -v`
Expected: FAIL — `undefined: Normalize`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/feeds/normalize.go
// Package feeds syncs ThreatWinds indicator lists into the mirror tree in the
// exact wire format the agent's netblock feed client consumes.
package feeds

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"errors"
	"io"
	"sort"
	"strings"
)

// Normalize converts an upstream accumulative artifact — a gzip'd tarball of
// member lists, or an already-gzip'd plain list — into the agent wire format:
// a gzip'd, newline-separated, deduped, sorted value list. Sorting makes the
// output bytes (and therefore the .sha256 sibling) deterministic.
func Normalize(upstream []byte) ([]byte, int, error) {
	gr, err := gzip.NewReader(bytes.NewReader(upstream))
	if err != nil {
		return nil, 0, errors.New("upstream artifact is not gzip")
	}
	raw, err := io.ReadAll(gr)
	if err != nil {
		return nil, 0, err
	}

	// A tarball's members are concatenated; anything else is one plain list.
	var text strings.Builder
	tr := tar.NewReader(bytes.NewReader(raw))
	isTar := false
	for {
		hdr, err := tr.Next()
		if err != nil {
			break
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		isTar = true
		b, err := io.ReadAll(tr)
		if err != nil {
			return nil, 0, err
		}
		text.Write(b)
		text.WriteByte('\n')
	}
	if !isTar {
		text.Reset()
		text.Write(raw)
	}

	seen := map[string]struct{}{}
	var values []string
	for _, line := range strings.Split(text.String(), "\n") {
		v := strings.TrimSpace(line)
		if v == "" || strings.HasPrefix(v, "#") {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		values = append(values, v)
	}
	sort.Strings(values)

	var out bytes.Buffer
	gw := gzip.NewWriter(&out)
	for _, v := range values {
		gw.Write([]byte(v))
		gw.Write([]byte{'\n'})
	}
	if err := gw.Close(); err != nil {
		return nil, 0, err
	}
	return out.Bytes(), len(values), nil
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/feeds/ -v`
Expected: PASS (3 tests).

---

### Task 5: `feeds` — upstream client + backend credential fetch

**Files:**
- Create: `utmstack-v12/edr/internal/feeds/client.go`, `utmstack-v12/edr/internal/feeds/creds.go`
- Test: `utmstack-v12/edr/internal/feeds/client_test.go`

**Interfaces:**
- Produces:
  - `type Client struct { BaseURL, APIKey, APISecret string; HTTP *http.Client }`
  - `func (c *Client) FetchAccumulative(ctx context.Context, level, name string) ([]byte, error)` — `GET {BaseURL}/feeds/v1/download/list/{level}/accumulative/{name}` with headers `api-key`/`api-secret`; non-200 is an error.
  - `func FetchCredentials(ctx context.Context, utmHost, internalKey string, httpc *http.Client) (apiKey, apiSecret string, err error)` — reads `utmstack.tw.apiKey` and `utmstack.tw.apiSecret` from `GET {utmHost}/api/v1/config/<key>` with header `X-Internal-Key` (response `{"value":"…"}`) — the same appconfig store `plugins/feeds` uses; the backend owns ThreatWinds registration, this daemon only reads.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/feeds/client_test.go
package feeds

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchAccumulativeSendsAuthHeadersAndPath(t *testing.T) {
	var gotPath, gotKey, gotSecret string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotKey = r.Header.Get("api-key")
		gotSecret = r.Header.Get("api-secret")
		w.Write([]byte("payload"))
	}))
	defer srv.Close()

	c := &Client{BaseURL: srv.URL, APIKey: "k", APISecret: "s", HTTP: srv.Client()}
	b, err := c.FetchAccumulative(context.Background(), "level1", "ip")
	if err != nil || string(b) != "payload" {
		t.Fatalf("body = %q, err %v", b, err)
	}
	if gotPath != "/feeds/v1/download/list/level1/accumulative/ip" {
		t.Fatalf("path = %s", gotPath)
	}
	if gotKey != "k" || gotSecret != "s" {
		t.Fatalf("auth headers = %q/%q", gotKey, gotSecret)
	}
}

func TestFetchAccumulativeErrorsOnNon200(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	c := &Client{BaseURL: srv.URL, HTTP: srv.Client()}
	if _, err := c.FetchAccumulative(context.Background(), "level1", "ip"); err == nil {
		t.Fatal("expected error on 503")
	}
}

func TestFetchCredentialsReadsBackendConfigKeys(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Internal-Key") != "ik" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch r.URL.Path {
		case "/api/v1/config/utmstack.tw.apiKey":
			w.Write([]byte(`{"value":"the-key"}`))
		case "/api/v1/config/utmstack.tw.apiSecret":
			w.Write([]byte(`{"value":"the-secret"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()
	k, s, err := FetchCredentials(context.Background(), srv.URL, "ik", srv.Client())
	if err != nil || k != "the-key" || s != "the-secret" {
		t.Fatalf("k=%q s=%q err=%v", k, s, err)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/feeds/ -run 'Fetch' -v`
Expected: FAIL — `undefined: Client` / `FetchCredentials`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/feeds/client.go
package feeds

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// Client downloads indicator lists from the ThreatWinds Feeds API.
type Client struct {
	BaseURL   string
	APIKey    string
	APISecret string
	HTTP      *http.Client
}

func (c *Client) FetchAccumulative(ctx context.Context, level, name string) ([]byte, error) {
	url := fmt.Sprintf("%s/feeds/v1/download/list/%s/accumulative/%s", c.BaseURL, level, name)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("api-key", c.APIKey)
	req.Header.Set("api-secret", c.APISecret)
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch %s: status %d", url, resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}
```

```go
// utmstack-v12/edr/internal/feeds/creds.go
package feeds

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// Backend appconfig keys where the platform stores its ThreatWinds
// credentials (registered and written by the backend / plugins-feeds flow —
// this daemon only reads them).
const (
	twAPIKeyKey    = "utmstack.tw.apiKey"
	twAPISecretKey = "utmstack.tw.apiSecret"
)

// FetchCredentials reads the ThreatWinds api-key/api-secret from the backend
// appconfig API, authenticated with the stack INTERNAL_KEY.
func FetchCredentials(ctx context.Context, utmHost, internalKey string, httpc *http.Client) (string, string, error) {
	get := func(key string) (string, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet,
			fmt.Sprintf("%s/api/v1/config/%s", utmHost, key), nil)
		if err != nil {
			return "", err
		}
		req.Header.Set("X-Internal-Key", internalKey)
		resp, err := httpc.Do(req)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return "", fmt.Errorf("config %s: status %d", key, resp.StatusCode)
		}
		var out struct {
			Value string `json:"value"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return "", err
		}
		return out.Value, nil
	}
	k, err := get(twAPIKeyKey)
	if err != nil {
		return "", "", err
	}
	s, err := get(twAPISecretKey)
	if err != nil {
		return "", "", err
	}
	return k, s, nil
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/feeds/ -v`
Expected: PASS (all feeds tests).

---

### Task 6: `feeds` — syncer (fetch → normalize → mirror, per-artifact isolation)

**Files:**
- Create: `utmstack-v12/edr/internal/feeds/sync.go`
- Test: `utmstack-v12/edr/internal/feeds/sync_test.go`

**Interfaces:**
- Consumes: `Normalize` (Task 4), `mirror.WriteArtifact` + `mirror.Recorder` (Tasks 2–3).
- Produces:
  ```go
  type Syncer struct {
      Root   string
      Levels []int    // e.g. [1]
      Names  []string // e.g. ["ip","domain","hostname"]
      Fetch  func(ctx context.Context, level, name string) ([]byte, error) // Client.FetchAccumulative in prod
      Rec    *mirror.Recorder
  }
  func (s *Syncer) SyncOnce(ctx context.Context)
  ```
  For each `(level, name)`: fetch → normalize → `WriteArtifact(Root, "feeds/v1/download/list/level{N}/accumulative/{name}", …)` → `Rec.FeedSuccess("level{N}/{name}", count, sha)`. Any failure: `Rec.FeedFailure` + continue with the next pair; the previous artifact is never touched.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/feeds/sync_test.go
package feeds

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

func TestSyncOnceWritesArtifactsAndIsolatesFailures(t *testing.T) {
	root := t.TempDir()
	s := &Syncer{
		Root:   root,
		Levels: []int{1},
		Names:  []string{"ip", "domain"},
		Rec:    mirror.NewRecorder(root),
		Fetch: func(_ context.Context, level, name string) ([]byte, error) {
			if name == "domain" {
				return nil, errors.New("upstream down")
			}
			return tarGzBytes(t, map[string]string{"m": "1.1.1.1\n"}), nil
		},
	}
	s.SyncOnce(context.Background())

	ipPath := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "accumulative", "ip")
	if _, err := os.Stat(ipPath); err != nil {
		t.Fatalf("ip artifact missing: %v", err)
	}
	if _, err := os.Stat(ipPath + ".sha256"); err != nil {
		t.Fatalf("ip checksum missing: %v", err)
	}
	domainPath := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "accumulative", "domain")
	if _, err := os.Stat(domainPath); !os.IsNotExist(err) {
		t.Fatal("failed feed must not produce an artifact")
	}
}

func TestSyncOnceFailureKeepsPreviousArtifact(t *testing.T) {
	root := t.TempDir()
	good := tarGzBytes(t, map[string]string{"m": "1.1.1.1\n"})
	fail := false
	s := &Syncer{
		Root: root, Levels: []int{1}, Names: []string{"ip"},
		Rec: mirror.NewRecorder(root),
		Fetch: func(_ context.Context, _, _ string) ([]byte, error) {
			if fail {
				return nil, errors.New("down")
			}
			return good, nil
		},
	}
	s.SyncOnce(context.Background())
	fail = true
	s.SyncOnce(context.Background())
	p := filepath.Join(root, "feeds", "v1", "download", "list", "level1", "accumulative", "ip")
	if _, err := os.Stat(p); err != nil {
		t.Fatal("previous artifact must survive a failed sync")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/feeds/ -run SyncOnce -v`
Expected: FAIL — `undefined: Syncer`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/feeds/sync.go
package feeds

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

// Syncer refreshes the mirrored indicator lists. Each (level, name) artifact
// is independent: one failure never blocks the others and never removes the
// previous good artifact (stale, never empty).
type Syncer struct {
	Root   string
	Levels []int
	Names  []string
	Fetch  func(ctx context.Context, level, name string) ([]byte, error)
	Rec    *mirror.Recorder
}

func (s *Syncer) SyncOnce(ctx context.Context) {
	for _, lvl := range s.Levels {
		level := fmt.Sprintf("level%d", lvl)
		for _, name := range s.Names {
			key := fmt.Sprintf("%s/%s", level, name)
			upstream, err := s.Fetch(ctx, level, name)
			if err != nil {
				log.Printf("feed %s: fetch failed: %v", key, err)
				s.Rec.FeedFailure(key, err.Error())
				continue
			}
			artifact, count, err := Normalize(upstream)
			if err != nil {
				log.Printf("feed %s: normalize failed: %v", key, err)
				s.Rec.FeedFailure(key, err.Error())
				continue
			}
			rel := fmt.Sprintf("feeds/v1/download/list/%s/accumulative/%s", level, name)
			if err := mirror.WriteArtifact(s.Root, rel, artifact); err != nil {
				log.Printf("feed %s: write failed: %v", key, err)
				s.Rec.FeedFailure(key, err.Error())
				continue
			}
			sum := sha256.Sum256(artifact)
			s.Rec.FeedSuccess(key, count, hex.EncodeToString(sum[:]))
			log.Printf("feed %s: %d indicators mirrored", key, count)
		}
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/feeds/ -v`
Expected: PASS (all feeds tests).

---

### Task 7: `sigs` — cvdupdate wrapper + CVD version reader

**Files:**
- Create: `utmstack-v12/edr/internal/sigs/sigs.go`
- Test: `utmstack-v12/edr/internal/sigs/sigs_test.go`

**Interfaces:**
- Consumes: `mirror.Recorder` (Task 3).
- Produces:
  ```go
  type Syncer struct {
      DBDir string                                             // /mirror/signatures
      Run   func(name string, args ...string) (string, error)  // default: exec.Command CombinedOutput
      Rec   *mirror.Recorder
  }
  func (s *Syncer) EnsureConfig() error   // cvd config set --dbdir <DBDir>
  func (s *Syncer) SyncOnce() error       // cvd update; on success Rec.SigSuccess(DatabaseVersions(DBDir))
  func DatabaseVersions(dir string) map[string]string  // {"daily.cvd":"27461", ...}
  func ReadCVDVersion(path string) (string, error)     // field 3 of the ClamAV-VDB header line
  ```
  CVD header format: the file starts with a 512-byte text header `ClamAV-VDB:<build time>:<version>:<num sigs>:…` — version is the colon-separated field at index 2.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/sigs/sigs_test.go
package sigs

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

func writeFakeCVD(t *testing.T, dir, name, version string) {
	t.Helper()
	header := "ClamAV-VDB:09 Jul 2026 10-00 -0400:" + version + ":100:63:...:sig:builder:1720000000"
	pad := make([]byte, 512-len(header))
	if err := os.WriteFile(filepath.Join(dir, name), append([]byte(header), pad...), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReadCVDVersion(t *testing.T) {
	dir := t.TempDir()
	writeFakeCVD(t, dir, "daily.cvd", "27461")
	v, err := ReadCVDVersion(filepath.Join(dir, "daily.cvd"))
	if err != nil || v != "27461" {
		t.Fatalf("version = %q, err %v", v, err)
	}
}

func TestDatabaseVersionsListsCVDs(t *testing.T) {
	dir := t.TempDir()
	writeFakeCVD(t, dir, "daily.cvd", "27461")
	writeFakeCVD(t, dir, "main.cvd", "62")
	_ = os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("x"), 0o644)
	dbs := DatabaseVersions(dir)
	if dbs["daily.cvd"] != "27461" || dbs["main.cvd"] != "62" || len(dbs) != 2 {
		t.Fatalf("dbs = %v", dbs)
	}
}

func TestSyncOnceRunsCvdUpdateAndRecords(t *testing.T) {
	dir := t.TempDir()
	writeFakeCVD(t, dir, "daily.cvd", "27461")
	var calls [][]string
	s := &Syncer{
		DBDir: dir,
		Rec:   mirror.NewRecorder(t.TempDir()),
		Run: func(name string, args ...string) (string, error) {
			calls = append(calls, append([]string{name}, args...))
			return "ok", nil
		},
	}
	if err := s.EnsureConfig(); err != nil {
		t.Fatal(err)
	}
	if err := s.SyncOnce(); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0][0] != "cvd" || calls[1][1] != "update" {
		t.Fatalf("calls = %v", calls)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/sigs/ -v`
Expected: FAIL — `undefined: Syncer` / `ReadCVDVersion`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/sigs/sigs.go
// Package sigs maintains the mirrored ClamAV signature databases via the
// official cvdupdate tool. cvdupdate is rate-limit-compliant against the CDN
// and maintains the exact CVD + cdiff layout freshclam PrivateMirror expects;
// content integrity is ClamAV's own signing, verified by the agents.
package sigs

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/utmstack/UTMStack/edr/internal/mirror"
)

type Syncer struct {
	DBDir string
	Run   func(name string, args ...string) (string, error)
	Rec   *mirror.Recorder
}

func New(dbDir string, rec *mirror.Recorder) *Syncer {
	return &Syncer{DBDir: dbDir, Rec: rec, Run: runCommand}
}

func runCommand(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).CombinedOutput()
	return string(out), err
}

// EnsureConfig points cvdupdate's database directory at the mirror tree.
func (s *Syncer) EnsureConfig() error {
	if err := os.MkdirAll(s.DBDir, 0o755); err != nil {
		return err
	}
	out, err := s.Run("cvd", "config", "set", "--dbdir", s.DBDir)
	if err != nil {
		return fmt.Errorf("cvd config: %v: %s", err, out)
	}
	return nil
}

// SyncOnce runs one cvdupdate cycle (full CVDs on first run, cdiffs after)
// and records the resulting database versions in status.json.
func (s *Syncer) SyncOnce() error {
	out, err := s.Run("cvd", "update")
	if err != nil {
		s.Rec.SigFailure(fmt.Sprintf("%v: %s", err, truncate(out, 500)))
		return fmt.Errorf("cvd update: %v", err)
	}
	s.Rec.SigSuccess(DatabaseVersions(s.DBDir))
	return nil
}

// DatabaseVersions maps each *.cvd in dir to its version (from the CVD header).
func DatabaseVersions(dir string) map[string]string {
	dbs := map[string]string{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return dbs
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cvd") {
			continue
		}
		if v, err := ReadCVDVersion(filepath.Join(dir, e.Name())); err == nil {
			dbs[e.Name()] = v
		}
	}
	return dbs
}

// ReadCVDVersion extracts the version from a CVD's 512-byte text header:
// "ClamAV-VDB:<build time>:<version>:...".
func ReadCVDVersion(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	head := make([]byte, 512)
	n, err := f.Read(head)
	if err != nil {
		return "", err
	}
	fields := strings.Split(string(head[:n]), ":")
	if len(fields) < 3 || fields[0] != "ClamAV-VDB" {
		return "", fmt.Errorf("%s: not a CVD header", path)
	}
	return fields[2], nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/sigs/ -v`
Expected: PASS (3 tests).

---

### Task 8: `daemon` — env config, sync loops, `main.go`

**Files:**
- Create: `utmstack-v12/edr/internal/daemon/daemon.go`
- Modify: `utmstack-v12/edr/main.go`
- Test: `utmstack-v12/edr/internal/daemon/daemon_test.go`

**Interfaces:**
- Consumes: `feeds.Client`, `feeds.FetchCredentials`, `feeds.Syncer`, `sigs.New`, `mirror.NewRecorder`.
- Produces:
  ```go
  type Config struct {
      MirrorDir            string        // EDR_MIRROR_DIR, default /mirror
      SigEvery             time.Duration // EDR_SIG_SYNC_MINUTES, default 60m
      FeedEvery            time.Duration // EDR_FEED_SYNC_HOURS, default 6h
      Levels               []int         // EDR_FEED_LEVELS, default [1]
      Names                []string      // EDR_FEED_NAMES, default ip,domain,hostname
      TWURL                string        // TW_API_URL, default https://apis.threatwinds.com
      TWKey, TWSecret      string        // TW_API_KEY / TW_API_SECRET (dev override)
      UTMHost              string        // UTM_HOST, default http://backend:8080
      InternalKey          string        // INTERNAL_KEY (required unless TW override set)
  }
  func FromEnv(getenv func(string) string) (Config, error)
  func Run(ctx context.Context, cfg Config) error
  ```
  `Run` wires the recorder + both syncers and drives two loops (immediate sync, then ticker). The feed loop first resolves credentials: env override wins; otherwise poll `FetchCredentials` once a minute until it succeeds (the backend owns registration).

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/edr/internal/daemon/daemon_test.go
package daemon

import (
	"testing"
	"time"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestFromEnvDefaults(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{"INTERNAL_KEY": "ik"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MirrorDir != "/mirror" || cfg.SigEvery != time.Hour || cfg.FeedEvery != 6*time.Hour {
		t.Fatalf("defaults = %+v", cfg)
	}
	if len(cfg.Levels) != 1 || cfg.Levels[0] != 1 {
		t.Fatalf("levels = %v", cfg.Levels)
	}
	if len(cfg.Names) != 3 || cfg.Names[0] != "ip" {
		t.Fatalf("names = %v", cfg.Names)
	}
	if cfg.UTMHost != "http://backend:8080" || cfg.TWURL == "" {
		t.Fatalf("hosts = %+v", cfg)
	}
}

func TestFromEnvOverridesAndValidation(t *testing.T) {
	cfg, err := FromEnv(env(map[string]string{
		"EDR_MIRROR_DIR": "/data", "EDR_SIG_SYNC_MINUTES": "30",
		"EDR_FEED_SYNC_HOURS": "2", "EDR_FEED_LEVELS": "1,2",
		"EDR_FEED_NAMES": "ip", "TW_API_KEY": "k", "TW_API_SECRET": "s",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MirrorDir != "/data" || cfg.SigEvery != 30*time.Minute || cfg.FeedEvery != 2*time.Hour {
		t.Fatalf("overrides = %+v", cfg)
	}
	if len(cfg.Levels) != 2 || cfg.Levels[1] != 2 || len(cfg.Names) != 1 {
		t.Fatalf("lists = %+v", cfg)
	}

	// Neither INTERNAL_KEY nor a TW override → error.
	if _, err := FromEnv(env(map[string]string{})); err == nil {
		t.Fatal("expected error when no credential source is configured")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./internal/daemon/ -v`
Expected: FAIL — `undefined: FromEnv`.

- [ ] **Step 3: Implement**

```go
// utmstack-v12/edr/internal/daemon/daemon.go
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
	"github.com/utmstack/UTMStack/edr/internal/sigs"
)

type Config struct {
	MirrorDir   string
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
```

Replace `utmstack-v12/edr/main.go`:
```go
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/utmstack/UTMStack/edr/internal/daemon"
)

func main() {
	cfg, err := daemon.FromEnv(os.Getenv)
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	log.Printf("EDR mirror daemon starting (mirror=%s)", cfg.MirrorDir)
	if err := daemon.Run(ctx, cfg); err != nil {
		log.Fatalf("run: %v", err)
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./... && go vet ./... && go build ./...`
Expected: all tests PASS, vet/build clean.

---

### Task 9: Dockerfile + local container acceptance

**Files:**
- Create: `utmstack-v12/edr/Dockerfile`
- Scratch (delete after): fixture-upstream server + fixture artifacts under the session scratchpad.

**Interfaces:**
- Consumes: the `edr` binary (Task 8). The Dockerfile matches the CI convention (Task 12): the reusable workflow builds a binary named `edr` inside `./edr`, then `docker build` with context `./edr`.

- [ ] **Step 1: Write the Dockerfile**

```dockerfile
FROM ubuntu:24.04

RUN apt-get update && \
    apt-get install -y ca-certificates python3 python3-pip && \
    pip3 install --break-system-packages cvdupdate && \
    apt-get clean && rm -rf /var/lib/apt/lists/*

COPY edr /app/edr

CMD ["/app/edr"]
```

- [ ] **Step 2: Build the binary and the image (Docker on the Mac is arm64)**

Run:
```bash
cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr
GOOS=linux GOARCH=arm64 go build -o edr .
docker build -t edr-mirror-test .
```
Expected: image builds. (`edr` here is the linux binary in the build context; remove it after the task — it must not linger in the repo tree.)

- [ ] **Step 3: Stand up a fixture ThreatWinds upstream on the host**

Create `<scratchpad>/twfixture/main.go`:
```go
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"log"
	"net/http"
)

func tarGz(lines string) []byte {
	var b bytes.Buffer
	gw := gzip.NewWriter(&b)
	tw := tar.NewWriter(gw)
	tw.WriteHeader(&tar.Header{Name: "list", Mode: 0o644, Size: int64(len(lines))})
	tw.Write([]byte(lines))
	tw.Close()
	gw.Close()
	return b.Bytes()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/feeds/v1/download/list/level1/accumulative/ip", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") != "test-key" {
			w.WriteHeader(401)
			return
		}
		w.Write(tarGz("1.1.1.1\n9.9.9.9\n"))
	})
	mux.HandleFunc("/feeds/v1/download/list/level1/accumulative/domain", func(w http.ResponseWriter, r *http.Request) {
		w.Write(tarGz("evil.example.com\n"))
	})
	mux.HandleFunc("/feeds/v1/download/list/level1/accumulative/hostname", func(w http.ResponseWriter, r *http.Request) {
		w.Write(tarGz("bad-host.example.com\n"))
	})
	log.Println("fixture upstream on :18080")
	log.Fatal(http.ListenAndServe(":18080", mux))
}
```
Run it in the background: `cd <scratchpad>/twfixture && go mod init twfixture && go run . &`

- [ ] **Step 4: Run the container and verify the mirror tree**

```bash
mkdir -p <scratchpad>/mirror
docker rm -f edr-mirror-test 2>/dev/null
docker run -d --name edr-mirror-test \
  -e TW_API_URL=http://host.docker.internal:18080 \
  -e TW_API_KEY=test-key -e TW_API_SECRET=test-secret \
  -v <scratchpad>/mirror:/mirror \
  edr-mirror-test
sleep 120 && docker logs edr-mirror-test
```
Verify (final outputs, not just logs):
1. `<scratchpad>/mirror/feeds/v1/download/list/level1/accumulative/ip` exists; `gunzip -c` shows `1.1.1.1` and `9.9.9.9` sorted; the `.sha256` sibling equals `shasum -a 256` of the artifact.
2. `<scratchpad>/mirror/signatures/` contains `main.cvd`, `daily.cvd`, `bytecode.cvd` (real CDN fetch — first run downloads ~300 MB; allow up to 15 minutes and poll `docker logs`, do not abort on a slow download).
3. `<scratchpad>/mirror/status.json` shows `signatures.last_success`, three `feeds` entries with counts, and no `last_error`.
4. Failure isolation: stop the fixture server, `docker restart edr-mirror-test`, wait one cycle: feed `last_error` set, `last_success` and artifacts preserved.

- [ ] **Step 5: Clean up**

```bash
docker rm -f edr-mirror-test
rm -rf <scratchpad>/twfixture <scratchpad>/mirror
rm -f /Users/atlas/UTMStack/EDR/utmstack-v12/edr/edr
```
(Keep the image if the VM acceptance in Task 18 is imminent; otherwise `docker rmi edr-mirror-test`.)

---

# Part B — platform wiring

### Task 10: `agentmanager` — serve `/private/edr` from the shared volume

**Files:**
- Modify: `utmstack-v12/agent-manager/config/global_const.go` (the `var (...)` block), `utmstack-v12/agent-manager/updates/updates.go`
- Test: `utmstack-v12/agent-manager/updates/updates_test.go`

**Interfaces:**
- Produces: `config.EDRMirrorFolder = "/edr-mirror"`; a package-private `newRouter() *gin.Engine` used by `ServeDependencies` and the test; route `GET /private/edr/*` → static files from `EDRMirrorFolder`.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/agent-manager/updates/updates_test.go
package updates

import (
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent-manager/config"
)

func TestEDRMirrorRouteServesStaticFiles(t *testing.T) {
	dir := t.TempDir()
	orig := config.EDRMirrorFolder
	config.EDRMirrorFolder = dir
	defer func() { config.EDRMirrorFolder = orig }()
	if err := os.WriteFile(filepath.Join(dir, "status.json"), []byte(`{"ok":true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	r := newRouter()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/private/edr/status.json", nil))
	if w.Code != 200 || w.Body.String() != `{"ok":true}` {
		t.Fatalf("code=%d body=%q", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest("GET", "/private/edr/missing", nil))
	if w.Code != 404 {
		t.Fatalf("missing file: code=%d", w.Code)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent-manager && go test ./updates/ -v`
Expected: FAIL — `undefined: newRouter` / `config.EDRMirrorFolder`.

- [ ] **Step 3: Implement**

In `config/global_const.go`, add to the existing `var (...)` block (after `UpdatesDependenciesFolder`):
```go
	// EDRMirrorFolder is the read-only mount of the EDR mirror volume
	// (signature databases + threat-intel feeds produced by the edr service).
	EDRMirrorFolder = "/edr-mirror"
```

In `updates/updates.go`, extract the router construction from `ServeDependencies` into:
```go
func newRouter() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(
		gin.Recovery(),
	)

	r.NoRoute(notFound)

	group := r.Group("/private")
	group.StaticFS("/dependencies", http.Dir(config.UpdatesDependenciesFolder))
	// EDR mirror tree, produced by the `edr` service into a shared volume.
	// If the volume isn't mounted (older installer) this simply 404s.
	group.StaticFS("/edr", http.Dir(config.EDRMirrorFolder))

	return r
}
```
and have `ServeDependencies` call `r := newRouter()` in place of the removed lines (the TLS/server code below it is unchanged).

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent-manager && go test ./updates/ -v && go build ./...`
Expected: PASS; module builds.

---

### Task 11: installer — `edr` service in the stack

**Files:**
- Modify: `utmstack-v12/installer/docker/stack.go`, `utmstack-v12/installer/docker/compose.go`

**Interfaces:**
- Produces: `StackConfig.EDRMirror string` (host dir `<DataDir>/edr-mirror`); compose service `edr`; `agentmanager` gains the read-only mirror mount; `Services` gains a memory-balance entry `{Name: "edr", Priority: 3, MinMemory: 150, MaxMemory: 512}`.

- [ ] **Step 1: stack.go**

Add to the `StackConfig` struct: `EDRMirror string`.
In `GetStackConfig()`, after the `stackConfig.ShmFolder = …` line:
```go
		stackConfig.EDRMirror = utils.MakeDir(0777, cnf.DataDir, "edr-mirror")
```
In the `Services` slice, add:
```go
			{Name: "edr", Priority: 3, MinMemory: 150, MaxMemory: 512},
```

- [ ] **Step 2: compose.go**

In `Populate`, add the service (after the `agentmanager` block, following its exact style):
```go
	edrMem := stack.ServiceResources["edr"].AssignedMemory
	c.Services["edr"] = Service{
		Image: utils.PointerOf[string]("ghcr.io/utmstack/utmstack/edr:${UTMSTACK_TAG}"),
		Volumes: []string{
			stack.EDRMirror + ":/mirror",
		},
		Environment: []string{
			"INTERNAL_KEY=" + conf.InternalKey,
			"UTM_HOST=http://backend:8080",
		},
		Logging: &dLogging,
		Deploy: &Deploy{
			Placement: &pManager,
			Resources: &Resources{
				Limits: &Res{
					Memory: utils.PointerOf[string](fmt.Sprintf("%vM", edrMem)),
				},
			},
		},
		DependsOn: []string{
			"backend",
		},
	}
```
And in the existing `agentmanager` service entry, append to `Volumes`:
```go
			stack.EDRMirror + ":/edr-mirror:ro",
```

- [ ] **Step 3: Verify**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/installer && go build ./... && go vet ./docker/`
Expected: clean. (The installer has no unit harness for `Populate`; full stack deploy is validated in the platform staging pipeline — out of scope here.)

---

### Task 12: CI — `build_edr` job

**Files:**
- Modify: `utmstack-v12/.github/workflows/v12-deployment-pipeline.yml`

**Interfaces:**
- Consumes: `.github/workflows/reusable-golang.yml` (runs `go test ./...` + `go build -o edr` in `./edr`, then `docker build` with default context `./edr` and dockerfile `./edr/Dockerfile` — exactly matching Tasks 8–9).

- [ ] **Step 1: Add the job**

Next to `build_backend` (same indentation level):
```yaml
  build_edr:
    name: Build EDR Mirror Microservice
    needs: [setup_deployment]
    if: ${{ needs.setup_deployment.outputs.tag != '' }}
    uses: ./.github/workflows/reusable-golang.yml
    with:
      image_name: edr
      tag: ${{ needs.setup_deployment.outputs.tag }}
```
Add `build_edr` to the `needs:` list of `all_builds_complete`.

- [ ] **Step 2: Validate the YAML**

Run: `python3 -c "import yaml; yaml.safe_load(open('/Users/atlas/UTMStack/EDR/utmstack-v12/.github/workflows/v12-deployment-pipeline.yml'))" && echo OK`
Expected: `OK`. (If PyYAML is unavailable in the venv, activate one and `pip install pyyaml` first, or validate with `docker run --rm -v "$PWD":/repo rhysd/actionlint:latest` — either check suffices.)

---

# Part C — agent-side client polish (module `utmstack-v12/agent/`)

### Task 13: netblock — auto-derived base URL + path-based checksum

**Files:**
- Modify: `utmstack-v12/agent/edr/config/config.go` (constants block near `EngineDir`), `utmstack-v12/agent/edr/netblock/feed.go`
- Test: `utmstack-v12/agent/edr/netblock/feed_urls_test.go` (new)

**Interfaces:**
- Produces:
  - `config.MirrorPort = "9001"`, `config.MirrorBasePath = "/private/edr"` (also consumed by Task 14).
  - `Feed.baseURL()` returns `MirrorBaseURL` (trimmed) when set; else `https://<Server>:9001/private/edr/feeds/v1`; else `""` when no server is registered — `UpdateOnce` then returns an error without fetching.
  - Checksum URL: `{base}/download/list/{level}/accumulative/{name}.sha256`.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/agent/edr/netblock/feed_urls_test.go
package netblock

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func gzList(t *testing.T, s string) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write([]byte(s)); err != nil {
		t.Fatal(err)
	}
	_ = w.Close()
	return b.Bytes()
}

func TestAutoDerivedURLsArePathBased(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.Blocklist.Levels = []int{1}
	cfg.Blocklist.IndicatorTypes = []string{"ip"}
	store := NewStore()
	f := NewFeed(cfg, c, store, nil)

	payload := gzList(t, "1.2.3.4\n")
	var fetched, checksummed []string
	f.Fetch = func(url string) (io.ReadCloser, error) {
		fetched = append(fetched, url)
		return io.NopCloser(bytes.NewReader(payload)), nil
	}
	f.FetchChecksum = func(url string) (string, error) {
		checksummed = append(checksummed, url)
		return sha(payload), nil
	}
	if err := f.UpdateOnce(context.Background()); err != nil {
		t.Fatal(err)
	}

	wantBase := "https://utm.example.com:9001/private/edr/feeds/v1"
	if len(fetched) != 1 || fetched[0] != wantBase+"/download/list/level1/accumulative/ip" {
		t.Fatalf("fetch urls = %v", fetched)
	}
	if len(checksummed) != 1 || checksummed[0] != wantBase+"/download/list/level1/accumulative/ip.sha256" {
		t.Fatalf("checksum urls = %v", checksummed)
	}
}

func TestUpdateOnceErrorsWithoutServerOrMirror(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.Server = ""
	store := NewStore()
	f := NewFeed(cfg, c, store, nil)
	called := false
	f.Fetch = func(string) (io.ReadCloser, error) { called = true; return nil, nil }
	if err := f.UpdateOnce(context.Background()); err == nil {
		t.Fatal("expected error when no mirror URL can be derived")
	}
	if called {
		t.Fatal("must not fetch without a mirror URL")
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/netblock/ -run 'AutoDerived|WithoutServer' -v`
Expected: FAIL (URL shape mismatch — the current code derives `<host>/feeds/v1` with no scheme and a `?level=` checksum URL).

- [ ] **Step 3: Implement**

In `edr/config/config.go`, alongside the existing path constants:
```go
	// Agent-facing mirror plane on the UTMStack server: the platform serves
	// the EDR mirror tree (signature databases + threat-intel feeds) as static
	// files on the dependencies port. Path segments deliberately avoid
	// engine/intel vendor names — these URLs surface in warning logs.
	MirrorPort     = "9001"
	MirrorBasePath = "/private/edr"
```

In `edr/netblock/feed.go`, replace `baseURL()`:
```go
func (f *Feed) baseURL() string {
	if f.cfg.Blocklist.MirrorBaseURL != "" {
		return strings.TrimRight(f.cfg.Blocklist.MirrorBaseURL, "/")
	}
	if f.cfg.Server == "" {
		return ""
	}
	return "https://" + f.cfg.Server + ":" + config.MirrorPort + config.MirrorBasePath + "/feeds/v1"
}
```
In `UpdateOnce`, compute the base once, before the loops:
```go
	base := f.baseURL()
	if base == "" {
		return fmt.Errorf("blocklist mirror not configured (no server registered)")
	}
```
and change the two URL lines inside the loop to:
```go
			url := fmt.Sprintf("%s/download/list/%s/accumulative/%s", base, level, typ)
			ckURL := fmt.Sprintf("%s/download/list/%s/accumulative/%s.sha256", base, level, typ)
```

- [ ] **Step 4: Run to verify it passes (plus no regressions)**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/netblock/ -v`
Expected: PASS — including all pre-existing feed tests (they inject `Fetch`/`FetchChecksum`, so the URL change is transparent to them; `config.Default()` has no `Server`, so any test relying on the old scheme-less fallback would have been failing fetches already — fix any that assert the old checksum URL shape by updating the expected string).

---

### Task 14: signature mirror semantics + `signature_fallback` config

**Files:**
- Modify: `utmstack-v12/agent/edr/config/config.go`, `utmstack-v12/agent/edr/engine/clamdconf.go`
- Test: `utmstack-v12/agent/edr/engine/clamdconf_test.go` (extend)

**Interfaces:**
- Consumes: `config.MirrorPort`/`config.MirrorBasePath` (Task 13).
- Produces:
  - `EDRConfig.SignatureFallback string` (`json:"signature_fallback"`, default `"cdn"`, values `"cdn"|"none"`).
  - `SigMirror` semantics: `""` (default) → auto-derive `https://<Server>:9001/private/edr/signatures`; `"cdn"` → official CDN only; any other value → explicit mirror URL.
  - `func ResolveMirrorURL(cfg config.EDRConfig) string` — the effective private-mirror URL (`""` = use the CDN).
  - `func RenderFreshclamConfCDN(cfg config.EDRConfig) string` — forces the official CDN (used by the failover cycle, Task 15).

- [ ] **Step 1: Write the failing test**

Add to `edr/engine/clamdconf_test.go`:
```go
func TestResolveMirrorURL(t *testing.T) {
	cfg := config.Default()
	if got := ResolveMirrorURL(cfg); got != "" {
		t.Fatalf("no server registered must resolve empty (CDN), got %q", got)
	}
	cfg.Server = "utm.example.com"
	want := "https://utm.example.com:9001/private/edr/signatures"
	if got := ResolveMirrorURL(cfg); got != want {
		t.Fatalf("auto = %q, want %q", got, want)
	}
	cfg.SigMirror = "cdn"
	if got := ResolveMirrorURL(cfg); got != "" {
		t.Fatalf("cdn sentinel must resolve empty, got %q", got)
	}
	cfg.SigMirror = "https://m.example.com/sigs"
	if got := ResolveMirrorURL(cfg); got != "https://m.example.com/sigs" {
		t.Fatalf("explicit = %q", got)
	}
}

func TestRenderFreshclamConfUsesResolvedMirror(t *testing.T) {
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	conf := RenderFreshclamConf(cfg)
	if !strings.Contains(conf, "PrivateMirror https://utm.example.com:9001/private/edr/signatures") {
		t.Fatalf("auto mirror missing:\n%s", conf)
	}
	cfg.SigMirror = "cdn"
	conf = RenderFreshclamConf(cfg)
	if !strings.Contains(conf, "DatabaseMirror database.clamav.net") || strings.Contains(conf, "PrivateMirror") {
		t.Fatalf("cdn sentinel must force the official CDN:\n%s", conf)
	}
}

func TestRenderFreshclamConfCDNForcesOfficial(t *testing.T) {
	cfg := config.Default()
	cfg.Server = "utm.example.com"
	cfg.SigMirror = "https://m.example.com/sigs"
	conf := RenderFreshclamConfCDN(cfg)
	if !strings.Contains(conf, "DatabaseMirror database.clamav.net") || strings.Contains(conf, "PrivateMirror") {
		t.Fatalf("CDN variant must ignore the mirror:\n%s", conf)
	}
}
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/engine/ -run 'ResolveMirror|Freshclam' -v`
Expected: FAIL — `undefined: ResolveMirrorURL` (and the auto-derive assertion).

- [ ] **Step 3: Implement**

`edr/config/config.go`:
- Update the `SigMirror` comment: `"" (default) auto-derives the UTMStack server mirror; "cdn" forces the official public database; any other value is an explicit mirror URL.`
- Add below `SigMirror`:
```go
	// SignatureFallback controls behavior when the private signature mirror is
	// unreachable: "cdn" (default) falls back to the official public database
	// for one cycle; "none" stays mirror-only (air-gapped posture).
	SignatureFallback string `json:"signature_fallback"`
```
- In `Default()`: `SignatureFallback: "cdn",`
- In `Load()` overlay (next to the `c.SigMirror = onDisk.SigMirror` line — note `SigMirror` is copied unconditionally so an explicit `"cdn"` survives; keep that):
```go
	if onDisk.SignatureFallback != "" {
		c.SignatureFallback = onDisk.SignatureFallback
	}
```

`edr/engine/clamdconf.go`:
```go
// ResolveMirrorURL returns the effective private-mirror URL for signature
// updates: "" ⇒ auto-derive from the registered UTMStack server; "cdn" ⇒
// none (use the official public database); anything else is explicit.
func ResolveMirrorURL(cfg config.EDRConfig) string {
	switch cfg.SigMirror {
	case "cdn":
		return ""
	case "":
		if cfg.Server == "" {
			return ""
		}
		return "https://" + cfg.Server + ":" + config.MirrorPort + config.MirrorBasePath + "/signatures"
	default:
		return cfg.SigMirror
	}
}

// RenderFreshclamConfCDN forces the official public database regardless of
// mirror config — used for a single fallback cycle when the mirror is down.
func RenderFreshclamConfCDN(cfg config.EDRConfig) string {
	forced := cfg
	forced.SigMirror = "cdn"
	return RenderFreshclamConf(forced)
}
```
In `RenderFreshclamConf`, replace the `if cfg.SigMirror != ""` block with:
```go
	if mirror := ResolveMirrorURL(cfg); mirror != "" {
		// Private mirror (the UTMStack server) — one upstream fetch fans out to
		// the fleet, avoids the official CDN's rate limits, and works on
		// restricted/air-gapped networks (§9).
		w("PrivateMirror " + mirror)
	} else {
		// Official public database (explicit "cdn", or no server registered).
		w("DatabaseMirror database.clamav.net")
	}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/engine/ ./edr/config/ -v`
Expected: PASS. If a pre-existing freshclam test asserts `DatabaseMirror` for a default config **with** a `Server` set, update it to the new auto-mirror expectation (the semantics change is deliberate and spec'd; nothing is shipped yet).

---

### Task 15: feed — mirror→CDN failover + faster error retry

**Files:**
- Modify: `utmstack-v12/agent/edr/feed/freshclam.go`, `utmstack-v12/agent/edr/engine/clamdconf.go` (one helper)
- Test: `utmstack-v12/agent/edr/feed/failover_test.go` (new)

**Interfaces:**
- Consumes: `engine.ResolveMirrorURL`, `engine.RenderFreshclamConf(CDN)` (Task 14).
- Produces:
  - `engine.WriteFreshclamConf(cfg config.EDRConfig, useCDN bool) error` — writes just freshclam.conf (mirror or CDN variant).
  - `Feed` fields `writeConf func(useCDN bool) error`, `mirrorFails int`, `onCDNFallback bool`; `const mirrorFailThreshold = 3`.
  - `func nextDelay(err error, interval time.Duration) time.Duration` — full interval on success; 15 min (capped at interval) after a failure, so failover is reachable in ~45 min instead of 3×`sig_update_hours`.
  - Task 17 reads `onCDNFallback` to decide the updater's TLS env.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/agent/edr/feed/failover_test.go
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
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/feed/ -run 'Failover|NoFailover|NextDelay' -v`
Expected: FAIL — undefined `writeConf`/`mirrorFailThreshold`/`nextDelay`.

- [ ] **Step 3: Implement**

`edr/engine/clamdconf.go` — add:
```go
// WriteFreshclamConf writes freshclam.conf, optionally forcing the official
// public database for a fallback cycle while the private mirror is down.
func WriteFreshclamConf(cfg config.EDRConfig, useCDN bool) error {
	conf := RenderFreshclamConf(cfg)
	if useCDN {
		conf = RenderFreshclamConfCDN(cfg)
	}
	return os.WriteFile(filepath.Join(config.EngineDir, "freshclam.conf"), []byte(conf), 0o644)
}
```

`edr/feed/freshclam.go`:
```go
const mirrorFailThreshold = 3

// add fields to Feed:
	writeConf     func(useCDN bool) error
	mirrorFails   int
	onCDNFallback bool

// in New():
	f.writeConf = func(useCDN bool) error { return engine.WriteFreshclamConf(cfg, useCDN) }
```
Replace `updateOnce`:
```go
func (f *Feed) updateOnce() (string, error) {
	err := f.runUpdater()
	if err != nil {
		logger.Error("UTMStack EDR: signature update failed: %v", err)
		if engine.ResolveMirrorURL(f.cfg) != "" && f.cfg.SignatureFallback == "cdn" {
			f.mirrorFails++
			if f.mirrorFails >= mirrorFailThreshold && !f.onCDNFallback {
				logger.Info("UTMStack EDR: signature mirror unreachable; falling back to public updates for one cycle")
				if werr := f.writeConf(true); werr == nil {
					f.onCDNFallback = true
					err = f.runUpdater()
				}
			}
		}
		if err != nil {
			return "", err
		}
	}
	if f.onCDNFallback {
		// Restore the mirror config so the next cycle retries the mirror.
		_ = f.writeConf(false)
		f.onCDNFallback = false
	}
	f.mirrorFails = 0
	ver, verr := f.currentSigDB()
	if verr != nil {
		return "", verr
	}
	if ver != "" {
		n, err := f.cache.MarkStaleBySigDB(ver)
		if err != nil {
			return ver, err
		}
		logger.Info("UTMStack EDR: signatures updated to %s, %d cached entries invalidated", ver, n)
	}
	return ver, nil
}

// nextDelay picks the wait before the next update attempt: the configured
// interval on success, a short retry after a failure (so mirror failover is
// reachable in minutes, not multiples of sig_update_hours).
func nextDelay(err error, interval time.Duration) time.Duration {
	const retry = 15 * time.Minute
	if err != nil && retry < interval {
		return retry
	}
	return interval
}
```
Rework `Run` to use it (replaces the ticker):
```go
func (f *Feed) Run(ctx context.Context) {
	interval := time.Duration(f.cfg.SigUpdateHours) * time.Hour
	if interval <= 0 {
		interval = 4 * time.Hour
	}
	_, err := f.updateOnce()
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(nextDelay(err, interval)):
			_, err = f.updateOnce()
		}
	}
}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/feed/ -v`
Expected: PASS (existing + new).

---

### Task 16: feed — freshness health event + status fields

**Files:**
- Modify: `utmstack-v12/agent/edr/feed/freshclam.go`, `utmstack-v12/agent/edr/service/service.go`
- Create: `utmstack-v12/agent/edr/feed/health.go`
- Test: `utmstack-v12/agent/edr/feed/health_test.go`

**Interfaces:**
- Consumes: `engine.ResolveMirrorURL` (Task 14), `event.Event`/`event.SourceEngine`/`event.ActionHealth`/`(*event.Spool).Append` (existing).
- Produces:
  - `func SignatureSource(cfg config.EDRConfig) string` → `"utmstack-mirror"` | `"official-cdn"` (branded — never the engine name).
  - `func (f *Feed) SetSpool(sp *event.Spool)`; `func (f *Feed) LastSuccess() time.Time`; `func (f *Feed) Stale(now time.Time, max time.Duration) bool` (mutex-guarded — status reads cross-goroutine).
  - One `health` event when staleness (> 3× interval) is first crossed; re-armed on the next success.
  - `statusDoc` gains `SignatureSource string json:"signature_source"` and `SignatureLastUpdate string json:"signature_last_update"`; `program` gains `sigFeed *feed.Feed`.

- [ ] **Step 1: Write the failing test**

```go
// utmstack-v12/agent/edr/feed/health_test.go
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
```

- [ ] **Step 2: Run to verify it fails**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/feed/ -run 'SignatureSource|Staleness' -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

`edr/feed/health.go`:
```go
package feed

import (
	"time"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/event"
)

// SignatureSource names the effective signature-update source for status and
// telemetry. Branded: never names the engine or the public database vendor.
func SignatureSource(cfg config.EDRConfig) string {
	if engine.ResolveMirrorURL(cfg) != "" {
		return "utmstack-mirror"
	}
	return "official-cdn"
}

func (f *Feed) SetSpool(sp *event.Spool) { f.spool = sp }

func (f *Feed) setLastSuccess(t time.Time) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.lastSuccess = t
}

// LastSuccess returns the time of the last successful signature update (zero
// if none this process lifetime). Safe for cross-goroutine status reads.
func (f *Feed) LastSuccess() time.Time {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.lastSuccess
}

// Stale reports whether the last success is older than max. A feed that has
// never succeeded has no baseline and is not reported stale.
func (f *Feed) Stale(now time.Time, max time.Duration) bool {
	ls := f.LastSuccess()
	if ls.IsZero() {
		return false
	}
	return now.Sub(ls) > max
}

// noteHealth emits one branded health event when staleness is first crossed;
// re-armed by the next success.
func (f *Feed) noteHealth(interval time.Duration) {
	if f.spool == nil {
		return
	}
	if !f.Stale(time.Now(), 3*interval) {
		return
	}
	if f.staleNotified {
		return
	}
	f.staleNotified = true
	ev := event.Event{
		Source:    event.SourceEngine,
		Action:    event.ActionHealth,
		Signature: "signature_updates_stale",
		Severity:  "warning",
	}
	if js, err := ev.ToJSON(); err == nil {
		_ = f.spool.Append(js)
	}
}
```

`edr/feed/freshclam.go` — add fields and hooks:
```go
// add to Feed struct:
	mu            sync.Mutex
	lastSuccess   time.Time
	spool         *event.Spool
	staleNotified bool
```
At the end of the success path in `updateOnce` (right before `return ver, nil`, after `f.mirrorFails = 0`):
```go
	f.setLastSuccess(time.Now())
	f.staleNotified = false
```
In `Run`, after each `updateOnce` call (both the initial one and the in-loop one), add:
```go
	f.noteHealth(interval)
```

`edr/service/service.go`:
- `program` struct: add `sigFeed *feed.Feed` (next to `netblock`).
- In `startPipeline`, replace `goSafe("feed", func() { feed.New(cfg, c).Run(ctx) })` with:
```go
	p.sigFeed = feed.New(cfg, c)
	p.sigFeed.SetSpool(sp) // sp = the spool created earlier in startPipeline
	goSafe("feed", func() { p.sigFeed.Run(ctx) })
```
(If the spool variable in `startPipeline` has a different name, use that name.)
- `statusDoc`: add
```go
	SignatureSource     string `json:"signature_source"`
	SignatureLastUpdate string `json:"signature_last_update"`
```
- `writeStatus` signature gains `sf *feed.Feed` (update the three call sites: initial + ticker in `run()`, plus any others `grep -n "writeStatus(" service.go` finds); populate:
```go
	doc.SignatureSource = feed.SignatureSource(cfg)
	if sf != nil {
		if t := sf.LastSuccess(); !t.IsZero() {
			doc.SignatureLastUpdate = t.UTC().Format(time.RFC3339)
		}
	}
```

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/feed/ ./edr/service/ -v`
Expected: PASS.

---

### Task 17: engine/feed — mirror CA bundle + env-aware updater

> **⚠️ SUPERSEDED (2026-07-08) — this task was implemented then REVERTED.** VM acceptance proved the ClamAV Windows build's libcurl validates against the Windows cert store and **ignores `CURL_CA_BUNDLE`**, so the CA-bundle mechanism below cannot work. It was replaced by a **flag-driven transport** (see the revised spec §7.4): `engine.ResolveMirrorURL` keys off `skip_cert_validate` — `false` → `https://<server>:9001/private/edr/signatures` (freshclam validates via the OS trust store natively), `true` → `http://<server>:9002/private/edr/signatures` (served plain-HTTP by the `edr` container; CVD integrity is by ClamAV signature, not transport). No Windows-root-store modification. `cabundle.go`/`cabundle_test.go`/`env_test.go` were deleted. The task text below is retained for history only — **do not implement it.**

**Files:**
- Create: `utmstack-v12/agent/edr/engine/cabundle.go`
- Modify: `utmstack-v12/agent/edr/feed/freshclam.go`
- Test: `utmstack-v12/agent/edr/engine/cabundle_test.go`, `utmstack-v12/agent/edr/feed/env_test.go`

**Interfaces:**
- Consumes: `engine.ResolveMirrorURL` (Task 14), `Feed.onCDNFallback` (Task 15).
- Produces:
  - `func MirrorHostPort(mirrorURL string) string` — `host:port` for an `https://` URL (default port 443), `""` otherwise.
  - `func CaptureMirrorCert(hostport string) (string, error)` — dials TLS (skip-verify), writes the presented chain as PEM to `EngineDir/mirror-ca.pem`, returns the path. TOFU-style pinning, consistent with the agent's `SkipCertValidate` posture; the signature updater itself has no skip-verify, hence the bundle.
  - `func (f *Feed) updaterEnv() []string` — `nil` normally; `CURL_CA_BUNDLE=<bundle>` appended to `os.Environ()` only when a bundle was captured **and** the current cycle targets the mirror (`!onCDNFallback`) — a CDN cycle must use the system trust store.

- [ ] **Step 1: Write the failing tests**

```go
// utmstack-v12/agent/edr/engine/cabundle_test.go
package engine

import (
	"crypto/x509"
	"encoding/pem"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestMirrorHostPort(t *testing.T) {
	cases := map[string]string{
		"https://utm.example.com:9001/private/edr/signatures": "utm.example.com:9001",
		"https://utm.example.com/sigs":                        "utm.example.com:443",
		"http://utm.example.com/sigs":                         "",
		"":                                                    "",
		"::bad::":                                             "",
	}
	for in, want := range cases {
		if got := MirrorHostPort(in); got != want {
			t.Fatalf("MirrorHostPort(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestCaptureMirrorCertWritesPEMChain(t *testing.T) {
	srv := httptest.NewTLSServer(nil)
	defer srv.Close()
	// Point EngineDir at a temp dir for the test.
	origEngineDir := config.EngineDir
	config.EngineDir = t.TempDir()
	defer func() { config.EngineDir = origEngineDir }()

	hostport := strings.TrimPrefix(srv.URL, "https://")
	p, err := CaptureMirrorCert(hostport)
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(b)
	if block == nil || block.Type != "CERTIFICATE" {
		t.Fatal("expected a PEM CERTIFICATE block")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	if !cert.Equal(srv.Certificate()) {
		t.Fatal("captured cert must match the server's")
	}
}
```
**Note:** `config.EngineDir` is currently a `const`/derived path — if it is not assignable, change it (and its siblings) from `const` to `var` in `edr/config/config.go`, or give `CaptureMirrorCert` the target dir as a parameter (`CaptureMirrorCert(hostport, dir string)`) and pass `config.EngineDir` at the call site; prefer the parameter form if `EngineDir` is const — then the test passes `t.TempDir()` directly and the production call passes `config.EngineDir`. Adjust the test accordingly; the produced interface below assumes the parameter form is NOT needed only if `EngineDir` is already a `var`.

```go
// utmstack-v12/agent/edr/feed/env_test.go
package feed

import (
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
)

func TestUpdaterEnvOnlySetForMirrorCycles(t *testing.T) {
	c, _ := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	defer c.Close()
	cfg := config.Default()
	cfg.SigMirror = "https://m.example.com/sigs"
	f := New(cfg, c)

	if env := f.updaterEnv(); env != nil {
		t.Fatal("no bundle captured → nil env")
	}
	f.caBundle = "/x/mirror-ca.pem"
	env := f.updaterEnv()
	found := false
	for _, e := range env {
		if e == "CURL_CA_BUNDLE=/x/mirror-ca.pem" {
			found = true
		}
	}
	if !found {
		t.Fatalf("bundle env missing: %v", env)
	}
	f.onCDNFallback = true
	if env := f.updaterEnv(); env != nil {
		t.Fatal("CDN fallback cycle must use the system trust store (nil env)")
	}
}
```

- [ ] **Step 2: Run to verify they fail**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/engine/ -run 'MirrorHostPort|CaptureMirror' -v && go test ./edr/feed/ -run UpdaterEnv -v`
Expected: FAIL — undefined.

- [ ] **Step 3: Implement**

`edr/engine/cabundle.go`:
```go
package engine

import (
	"bytes"
	"crypto/tls"
	"encoding/pem"
	"net/url"
	"os"
	"path/filepath"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

// MirrorHostPort extracts host:port from an https mirror URL ("" for
// anything else — a plain-http mirror needs no CA bundle).
func MirrorHostPort(mirrorURL string) string {
	if mirrorURL == "" {
		return ""
	}
	u, err := url.Parse(mirrorURL)
	if err != nil || u.Scheme != "https" || u.Hostname() == "" {
		return ""
	}
	port := u.Port()
	if port == "" {
		port = "443"
	}
	return u.Hostname() + ":" + port
}

// CaptureMirrorCert connects to the mirror and writes its presented
// certificate chain to EngineDir/mirror-ca.pem so the signature updater
// (which has no skip-verify option) can trust the on-prem server's TLS cert.
// TOFU-style pinning, consistent with the agent's SkipCertValidate posture.
func CaptureMirrorCert(hostport string) (string, error) {
	conn, err := tls.Dial("tcp", hostport, &tls.Config{InsecureSkipVerify: true})
	if err != nil {
		return "", err
	}
	defer conn.Close()
	var buf bytes.Buffer
	for _, c := range conn.ConnectionState().PeerCertificates {
		_ = pem.Encode(&buf, &pem.Block{Type: "CERTIFICATE", Bytes: c.Raw})
	}
	p := filepath.Join(config.EngineDir, "mirror-ca.pem")
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		return "", err
	}
	return p, nil
}
```

`edr/feed/freshclam.go`:
- Add field `caBundle string` to `Feed`.
- At the top of `Run` (before the first `updateOnce`):
```go
	if u := engine.ResolveMirrorURL(f.cfg); u != "" {
		if hp := engine.MirrorHostPort(u); hp != "" {
			if p, err := engine.CaptureMirrorCert(hp); err == nil {
				f.caBundle = p
			} else {
				logger.Error("UTMStack EDR: could not capture mirror certificate: %v", err)
			}
		}
	}
```
- Add:
```go
// updaterEnv returns the environment for the signature-updater subprocess:
// the pinned mirror CA bundle for mirror cycles, the system default for CDN
// fallback cycles (the public database needs the system trust store).
func (f *Feed) updaterEnv() []string {
	if f.caBundle == "" || f.onCDNFallback {
		return nil
	}
	return append(os.Environ(), "CURL_CA_BUNDLE="+f.caBundle)
}
```
- Replace `defaultRunUpdater` (drops the `shared/exec` import in favor of a local env-capable runner):
```go
func (f *Feed) defaultRunUpdater() error {
	// Runs the signature updater bundled with the engine. Naming here is
	// internal only; nothing is surfaced in events/logs.
	bin := filepath.Join(config.EngineDir, "freshclam.exe")
	cmd := osexec.Command(bin, "--config-file", filepath.Join(config.EngineDir, "freshclam.conf"))
	cmd.Dir = config.EngineDir
	if env := f.updaterEnv(); env != nil {
		cmd.Env = env
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		if stderr.Len() > 0 {
			return fmt.Errorf("updater failed: %s: %w", stderr.String(), err)
		}
		return fmt.Errorf("updater failed: %w", err)
	}
	return nil
}
```
(import `osexec "os/exec"`, `"bytes"`, `"fmt"`, `"os"`; remove the now-unused `shared/exec` import if nothing else uses it.)

- [ ] **Step 4: Run to verify it passes**

Run: `cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent && go test ./edr/engine/ ./edr/feed/ -v`
Expected: PASS.

---

### Task 18: Full verification + VM acceptance (end-to-end against the real container)

**Files:**
- No new repo files. Scratch helpers (mirror TLS front) under the session scratchpad — delete afterwards.

- [ ] **Step 1: Whole-module verification (host)**

```bash
cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent
go test ./edr/... ./agent/
go test ./edr/netblock/ ./edr/feed/ -race
for arch in amd64 arm64; do
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_edr_windows_$arch.exe ./edr
  GOOS=windows GOARCH=$arch go build -o dist/utmstack_agent_windows_$arch.exe .
  GOOS=windows GOARCH=$arch go build ./...
  GOOS=windows GOARCH=$arch go vet ./edr/...
done
# labeling audit — BOTH greps must return nothing
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*clam' edr/
grep -rniE '(logger\.(Info|Error|Fatal|Debug)|fmt\.(Print|Sprint))[^\n]*threatwinds' edr/
cd /Users/atlas/UTMStack/EDR/utmstack-v12/edr && go test ./... && go vet ./...
cd /Users/atlas/UTMStack/EDR/utmstack-v12/agent-manager && go test ./updates/ && go build ./...
cd /Users/atlas/UTMStack/EDR/utmstack-v12/installer && go build ./...
```
Expected: everything green; audits print nothing.

- [ ] **Step 2: Run the `edr` container + a TLS static front on the Mac**

The VM reaches the Mac at `10.211.55.2`. Mimic the agentmanager serving plane:
```bash
mkdir -p <scratchpad>/vmmirror/mirror
docker rm -f edr-mirror-test 2>/dev/null
docker run -d --name edr-mirror-test \
  -e TW_API_KEY=... -e TW_API_SECRET=... \   # real creds if available; else re-use the Task 9 fixture upstream with TW_API_URL
  -v <scratchpad>/vmmirror/mirror:/mirror \
  edr-mirror-test
```
TLS front (`<scratchpad>/vmmirror/front/main.go`):
```go
package main

import (
	"log"
	"net/http"
)

func main() {
	mux := http.NewServeMux()
	mux.Handle("/private/edr/", http.StripPrefix("/private/edr/",
		http.FileServer(http.Dir("../mirror"))))
	log.Println("mirror front on :9001 (TLS)")
	log.Fatal(http.ListenAndServeTLS(":9001", "cert.pem", "key.pem", mux))
}
```
```bash
cd <scratchpad>/vmmirror/front
openssl req -x509 -newkey ec -pkeyopt ec_paramgen_curve:P-256 \
  -keyout key.pem -out cert.pem -days 7 -nodes -subj "/CN=10.211.55.2"
go mod init front && go run . &
curl -sk https://10.211.55.2:9001/private/edr/status.json   # sanity from the Mac
```

- [ ] **Step 3: Point the VM's EDR at it (fresh config, auto-derive)**

Follow the reliable-redeploy recipe (stop → taskkill → verify GONE → download over HTTP → hash-verify → start) to ship the freshly built `utmstack_edr_windows_arm64.exe`. Edit `edr.json` on the VM: ensure `"server": "10.211.55.2"`, **no** `signature_mirror`, **no** `blocklist_mirror` (auto-derive is the point), `"skip_cert_validate": true`. Restart the EDR service.

Verify (final outputs, not logs):
1. Generated `engine/freshclam.conf` contains `PrivateMirror https://10.211.55.2:9001/private/edr/signatures`; `engine/mirror-ca.pem` exists.
2. Signature update succeeds from the mirror (front's access = check the container/front logs; `edr-status` → `signature_source: "utmstack-mirror"`, recent `signature_last_update`, `sigdb_version` non-empty).
3. Tamper test: flip one byte mid-file in `<scratchpad>/vmmirror/mirror/signatures/daily.cvd`, force an update cycle (restart the service) → the updater **rejects** it (no new version loaded); restore the file.
4. `status.json` on the VM: `blocklist_indicators` > 0 pulled via the auto-derived feed URL. If using the fixture upstream, `1.1.1.1` is in the feed → from the VM `curl http://1.1.1.1 --connect-timeout 3` fails while `curl http://8.8.8.8 --connect-timeout 3` (control) is unaffected, and a branded `network_watcher` event reaches the spool/platform.
5. cdiff check: after the container's next signature cycle publishes a newer daily, the VM's next update fetches a `*.cdiff` (front log shows a `.cdiff` GET), not a full `daily.cvd`.

- [ ] **Step 4: Failover + staleness**

1. Kill the TLS front. Restart the EDR service (first mirror attempt fails; retry cadence is 15 min) → within ~45 min the log shows "signature mirror unreachable; falling back to public updates" and a successful CDN cycle; `signature_source` stays `utmstack-mirror` (source reflects config, the fallback is one cycle). If VM time is constrained: verify the 3-failure failover once via the unit-tested path plus a single observed mirror-failure log line, and note it.
2. Blocklist during the outage: `blocklist_feed_stale` flips per its threshold; indicators keep enforcing (stale, never empty).
3. Restart the front → next cycles return to the mirror (front log shows GETs again).

- [ ] **Step 5: Labeling audit on live output + cleanup**

1. **Stop the EDR service first** (a live telemetry box self-references audit commands), then grep the VM-side spool/exported events for `clam` and `threatwinds` — nothing may match (mirror URLs contain neither, by design).
2. Clean up: `docker rm -f edr-mirror-test`, kill the front, `rm -rf <scratchpad>/vmmirror`, remove any linux `edr` binary from the repo tree, restore any VM config changed for testing.
3. Record found/fixed bugs in the plan doc or memory per project convention.

---

## Self-Review (against the spec)

**Coverage:** §3.1 repo/image → Tasks 1, 9; §3.2 cvdupdate → Task 7; §3.3 feed sync/normalize/atomicity → Tasks 4–6; §3.4 credentials → Task 5 (+ daemon polling, Task 8); §3.5 status.json → Task 3; §3.6 env config → Task 8; §3.7 stale-never-empty → Tasks 3, 6 (tests); §4 agentmanager route → Task 10; §5 installer → Task 11; §6 CI → Task 12; §7.1 checksum path → Task 13; §7.2 auto-derive + defaults → Tasks 13–14; §7.3 failover/health/status → Tasks 15–16; §7.4 CA bundle → Task 17; §9 compat + §10 testing → Task 18. §7.5 exclusions honored (no daily deltas, no engine provisioning).

**Type consistency:** `mirror.Recorder` methods (`SigSuccess/SigFailure/FeedSuccess/FeedFailure`) used identically in Tasks 6–7; `engine.ResolveMirrorURL` signature consistent across Tasks 14–17; `Feed.onCDNFallback` produced in Task 15, consumed in Task 17; `config.MirrorPort`/`MirrorBasePath` produced in Task 13, consumed in Task 14.

**Known judgment points (flagged for the implementer, not placeholders):** the spool variable name in `startPipeline` (Task 16) and whether `config.EngineDir` is assignable in tests (Task 17) — both have explicit fallback instructions inline.

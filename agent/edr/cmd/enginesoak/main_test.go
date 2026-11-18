package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func corpusForTest(t *testing.T) (string, manifest) {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "files"), 0700); err != nil {
		t.Fatal(err)
	}
	m := manifest{Schema: 1}
	for _, entry := range []struct{ name, data string }{{"hit.fixture", "safe-unique-marker-20261020"}, {"benign.fixture", "ordinary harmless contents"}} {
		data := []byte(entry.data)
		sum := sha256.Sum256(data)
		if err := os.WriteFile(filepath.Join(root, "files", entry.name), data, 0600); err != nil {
			t.Fatal(err)
		}
		m.Files = append(m.Files, fixture{Path: entry.name, SHA256: hex.EncodeToString(sum[:]), Bytes: int64(len(data)), Label: "test", Family: "raw"})
	}
	raw, _ := json.Marshal(m)
	if err := os.WriteFile(filepath.Join(root, "manifest.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	return root, m
}

func TestManifestIdentityAndUnsafeFiles(t *testing.T) {
	root, m := corpusForTest(t)
	loaded, hash, err := loadManifest(root)
	if err != nil || len(loaded.Files) != 2 || len(hash) != 64 {
		t.Fatalf("load: %v %+v %s", err, loaded, hash)
	}
	for _, name := range []string{"../outside", "/etc/passwd", "./hit.fixture", "a/../hit.fixture", "a\\hit.fixture", "C:hit.fixture"} {
		if _, err := safeFile(filepath.Join(root, "files"), name); err == nil {
			t.Errorf("accepted unsafe %q", name)
		}
	}
	if err := os.Symlink(filepath.Join(root, "files", m.Files[0].Path), filepath.Join(root, "files", "linked.fixture")); err != nil {
		t.Fatal(err)
	}
	if _, err := safeFile(filepath.Join(root, "files"), "linked.fixture"); err == nil {
		t.Fatal("accepted symlink")
	}
	if err := os.WriteFile(filepath.Join(root, "files", m.Files[0].Path), []byte(strings.Repeat("x", int(m.Files[0].Bytes))), 0600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := loadManifest(root); err == nil {
		t.Fatal("accepted changed content")
	}
}

func TestManifestBoundsAndDuplicates(t *testing.T) {
	for _, mutation := range []func(*manifest){
		func(m *manifest) { m.Files = append(m.Files, m.Files[0]) },
		func(m *manifest) { m.Files[0].Bytes = -1 },
		func(m *manifest) { m.Files[0].Bytes = 201 << 20 },
		func(m *manifest) { m.Files[0].SHA256 = "bad" },
		func(m *manifest) { m.Files[0].Path = "../outside" },
		func(m *manifest) { m.Schema = 0 },
	} {
		root, m := corpusForTest(t)
		mutation(&m)
		raw, _ := json.Marshal(m)
		if err := os.WriteFile(filepath.Join(root, "manifest.json"), raw, 0600); err != nil {
			t.Fatal(err)
		}
		if _, _, err := loadManifest(root); err == nil {
			t.Fatal("accepted invalid manifest")
		}
	}
}

func TestOptionBounds(t *testing.T) {
	valid := options{DB: "db", Corpus: "corpus", Out: "new", Workers: 1, Rounds: 1, SampleMS: 100}
	if err := validateOptions(valid); err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*options){func(o *options) { o.Workers = 17 }, func(o *options) { o.Workers = 0 }, func(o *options) { o.Rounds = 101 }, func(o *options) { o.Reloads = -1 }, func(o *options) { o.Reloads = 101 }, func(o *options) { o.SampleMS = 1 }, func(o *options) { o.Limits.MaxFileSizeMB = 201 }} {
		o := valid
		mutate(&o)
		if err := validateOptions(o); err == nil {
			t.Fatal("accepted invalid options")
		}
	}
}

type fakeScanner struct {
	calls    atomic.Int64
	unstable bool
}

func (f *fakeScanner) Version() string { return "same-pinned-generation" }
func (f *fakeScanner) Reload() error   { time.Sleep(20 * time.Millisecond); return nil }
func (f *fakeScanner) ScanSnapshot(ctx context.Context, data []byte) (bool, string, string, error) {
	n := f.calls.Add(1)
	select {
	case <-ctx.Done():
		return false, "", f.Version(), ctx.Err()
	case <-time.After(5 * time.Millisecond):
	}
	if bytes.Contains(data, []byte("unique-marker")) || (f.unstable && n%2 == 1) {
		return false, "Safe.Test", f.Version(), nil
	}
	return true, "", f.Version(), nil
}

type overlappingScanner struct {
	fakeScanner
	ctx            context.Context
	started        chan struct{}
	released       chan struct{}
	scans, reloads atomic.Int64
}

func (f *overlappingScanner) ScanSnapshot(ctx context.Context, data []byte) (bool, string, string, error) {
	if f.scans.Add(1) <= 2 {
		f.started <- struct{}{}
		select {
		case <-f.released:
		case <-ctx.Done():
			return false, "", f.Version(), ctx.Err()
		}
	}
	return f.fakeScanner.ScanSnapshot(ctx, data)
}

func (f *overlappingScanner) Reload() error {
	n := f.reloads.Add(1)
	if n == 1 {
		for i := 0; i < 2; i++ {
			select {
			case <-f.started:
			case <-f.ctx.Done():
				return f.ctx.Err()
			}
		}
	}
	err := f.fakeScanner.Reload()
	if n == 3 {
		close(f.released)
	}
	return err
}

func TestResidentCoverageStabilityAndReloadOverlap(t *testing.T) {
	root, m := corpusForTest(t)
	// Hold both initial scan calls until the three reloads finish. Timer
	// ratios alone cannot guarantee overlap when the test machine is busy.
	o := options{Corpus: root, Workers: 2, Rounds: 32, Reloads: 3}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	scanner := &overlappingScanner{ctx: ctx, started: make(chan struct{}, 2), released: make(chan struct{})}
	var out bytes.Buffer
	count, total, counts, stable, maxActive, _, reloads, _, _, err := scanRounds(ctx, scanner, o, m, time.Now(), json.NewEncoder(&out), nil)
	if err != nil || count != 64 || !stable || maxActive < 2 || counts["detected"] != 32 || counts["clean"] != 32 {
		t.Fatalf("unexpected result: %d %v %v %d %v", count, counts, stable, maxActive, err)
	}
	if total != (m.Files[0].Bytes+m.Files[1].Bytes)*32 {
		t.Fatal("wrong bytes")
	}
	if len(reloads) != 3 {
		t.Fatal("missing reloads")
	}
	for _, r := range reloads {
		if r.OverlapNS <= 0 {
			t.Fatal("missing actual scan overlap")
		}
	}
	dec := json.NewDecoder(&out)
	seen := map[string]bool{}
	for i := 0; i < count; i++ {
		var r record
		if err := dec.Decode(&r); err != nil {
			t.Fatal(err)
		}
		key := string(rune(r.Round)) + r.Fixture
		if seen[key] {
			t.Fatal("duplicate coverage")
		}
		seen[key] = true
	}
}
func TestUnstableVerdictsAndCancellationAreVisible(t *testing.T) {
	root, m := corpusForTest(t)
	m.Files = m.Files[1:]
	var out bytes.Buffer
	_, _, _, stable, _, _, _, _, _, err := scanRounds(context.Background(), &fakeScanner{unstable: true}, options{Corpus: root, Workers: 1, Rounds: 4}, m, time.Now(), json.NewEncoder(&out), nil)
	if err != nil || stable {
		t.Fatalf("unstable hidden: stable=%v err=%v", stable, err)
	}
	if v := classify(false, "", context.Canceled); v.State != "incomplete" || v.Reason != "Timeout" {
		t.Fatal(v)
	}
	if v := classify(false, "Safe.Test", errors.New("partial")); v.State != "detected" {
		t.Fatal(v)
	}
}
func TestOverlapUsesUnionNotDoubleCounting(t *testing.T) {
	w := mergedWindows([]window{{10, 30}, {0, 20}, {50, 60}})
	if got := overlapNS(window{5, 55}, w); got != 30 {
		t.Fatalf("overlap %d", got)
	}
}

type brokenWriter struct{}

func (brokenWriter) Write([]byte) (int, error) { return 0, io.ErrClosedPipe }
func TestRecordWriteFailureIsReturned(t *testing.T) {
	root, m := corpusForTest(t)
	_, _, _, _, _, _, _, _, _, err := scanRounds(context.Background(), &fakeScanner{}, options{Corpus: root, Workers: 1, Rounds: 1}, m, time.Now(), json.NewEncoder(brokenWriter{}), nil)
	if !errors.Is(err, io.ErrClosedPipe) {
		t.Fatal(err)
	}
}
func TestSmallRuleEndToEnd(t *testing.T) {
	root, _ := corpusForTest(t)
	db := t.TempDir()
	if err := os.WriteFile(filepath.Join(db, "safe.ndb"), []byte("Safe.Test:0:*:"+hex.EncodeToString([]byte("safe-unique-marker-20261020"))+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "run")
	report, err := execute(context.Background(), options{DB: db, Corpus: root, Out: out, Workers: 1, Rounds: 3, Reloads: 2, SampleMS: 25})
	if err != nil {
		t.Fatal(err)
	}
	if !report.Valid || report.ForcedGC || report.ScanCount != 6 || report.Counts["detected"] != 3 || report.Counts["clean"] != 3 || len(report.Reloads) != 2 {
		t.Fatalf("%+v", report)
	}
	for _, name := range []string{"summary.json", "records.jsonl", "runtime.jsonl"} {
		if _, err := os.Stat(filepath.Join(out, name)); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = execute(context.Background(), options{DB: db, Corpus: root, Out: out, Workers: 1, Rounds: 1, SampleMS: 100}); err == nil {
		t.Fatal("overwrote old evidence")
	}
}

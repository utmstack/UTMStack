package engine

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
)

const testMarker = "UTMSTACK-NATIVE-HARMLESS-TEST"

func nativeTestEngine(t *testing.T) *Engine {
	t.Helper()
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "local.ndb"), []byte("Native.Test:0:*:"+hex.EncodeToString([]byte(testMarker))+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	return e
}
func zipFixture(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := zip.NewWriter(&b)
	f, err := w.Create("../../payload.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = f.Write(data); err != nil {
		t.Fatal(err)
	}
	if err = w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func gzipFixture(t *testing.T, data []byte) []byte {
	t.Helper()
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	if _, err := w.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	return b.Bytes()
}
func TestNativeRecursiveDetection(t *testing.T) {
	e := nativeTestEngine(t)
	var tarbuf bytes.Buffer
	tw := tar.NewWriter(&tarbuf)
	_ = tw.WriteHeader(&tar.Header{Name: "x", Mode: 0600, Size: int64(len(testMarker)), Format: tar.FormatUSTAR})
	_, _ = tw.Write([]byte(testMarker))
	_ = tw.Close()
	utf := []byte{0xff, 0xfe}
	for _, b := range []byte(testMarker) {
		utf = append(utf, b, 0)
	}
	mail := "MIME-Version: 1.0\r\nContent-Type: multipart/mixed; boundary=abc\r\n\r\n--abc\r\nContent-Type: application/octet-stream\r\nContent-Transfer-Encoding: base64\r\n\r\n" + base64.StdEncoding.EncodeToString([]byte(testMarker)) + "\r\n--abc--\r\n"
	for name, data := range map[string][]byte{"raw": []byte(testMarker), "zip": zipFixture(t, []byte(testMarker)), "nested": zipFixture(t, gzipFixture(t, []byte(testMarker))), "tar": tarbuf.Bytes(), "gzip": gzipFixture(t, tarbuf.Bytes()), "mail": []byte(mail)} {
		t.Run(name, func(t *testing.T) {
			clean, sig, version, err := e.ScanSnapshot(context.Background(), data)
			if err != nil || clean || sig != "Native.Test" || version != e.Version() {
				t.Fatalf("clean=%v sig=%q version=%q err=%v", clean, sig, version, err)
			}
		})
	}
	// UTF-16 text is matched raw and as normalized, lowercased text, as the
	// reference matches it; an upper-case ASCII rule matches neither.
	if clean, sig, _, err := e.ScanSnapshot(context.Background(), utf); !clean || sig != "" || err != nil {
		t.Fatalf("utf16: clean=%v sig=%q err=%v", clean, sig, err)
	}
}

// A ZIP archive without an end record: the reference reads nothing from it
// (docs/edr/native-engine/parity/zip.md), so only its raw bytes are matched.
func TestNativeZIPWithoutDirectoryIsDecided(t *testing.T) {
	e := nativeTestEngine(t)
	if clean, sig, err := e.Scan(context.Background(), []byte("PK\x03\x04broken")); !clean || sig != "" || err != nil {
		t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
	}
}

func TestNativeIncompleteNeverClean(t *testing.T) {
	e := nativeTestEngine(t)
	for name, data := range map[string][]byte{"pdf": []byte("%PDF-1.7..."), "rar": []byte("Rar!\x1a\x07x"), "gzip": {0x1f, 0x8b, 0, 0, 0, 0, 0, 0}} {
		t.Run(name, func(t *testing.T) {
			clean, sig, err := e.Scan(context.Background(), data)
			if clean || sig != "" || err == nil {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
	// With the default broken-executable heuristic enabled, a truncated DOS
	// header is a detection in the reference rather than an incomplete scan.
	if clean, sig, err := e.Scan(context.Background(), []byte("MZbroken")); clean || sig != "Heuristics.Broken.Executable" || err != nil {
		t.Fatalf("truncated PE: clean=%v sig=%q err=%v", clean, sig, err)
	}
	clean, _, err := e.Scan(context.Background(), []byte("ordinary benign content"))
	if !clean || err != nil {
		t.Fatal(clean, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if clean, _, err := e.Scan(ctx, []byte(testMarker)); clean || !errors.Is(err, context.Canceled) {
		t.Fatal(clean, err)
	}
}
func TestNativeArchiveBounds(t *testing.T) {
	e := nativeTestEngine(t)
	g := e.current.Load()
	data := gzipFixture(t, bytes.Repeat([]byte("a"), 8192))
	s := scanState{ctx: context.Background(), db: g.db, maxFile: 1024, remaining: 2048, maxFiles: 10, maxDepth: 2}
	if _, err := s.visit(data, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	nested := gzipFixture(t, gzipFixture(t, []byte("benign")))
	s = scanState{ctx: context.Background(), db: g.db, maxFile: 8192, remaining: 8192, maxFiles: 10, maxDepth: 1}
	if _, err := s.visit(nested, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
}
func manifestHash(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func TestNativeAtomicGeneration(t *testing.T) {
	e := nativeTestEngine(t)
	files := map[string][]byte{"x.ndb": []byte("Other.Test:0:*:414243444546\n")}
	v, err := e.ActivateGeneration(files, 2, manifestHash("2"))
	if err != nil {
		t.Fatal(err)
	}
	if seq, h := e.FeedState(); seq != 2 || h != manifestHash("2") {
		t.Fatal(seq, h)
	}
	if _, err := e.ActivateGeneration(files, 1, manifestHash("1")); err == nil {
		t.Fatal("rollback accepted")
	}
	if _, err := e.ActivateGeneration(files, 2, manifestHash("evil")); err == nil {
		t.Fatal("equivocation accepted")
	}
	if _, err := e.ActivateGeneration(map[string][]byte{"bad.ndb": []byte("broken")}, 3, manifestHash("3")); err == nil {
		t.Fatal("bad generation accepted")
	}
	if e.Version() != v {
		t.Fatal("failed update replaced generation")
	}
	reload := New(e.cfg)
	if err := reload.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	if reload.Version() != v {
		t.Fatal("restart lost generation")
	}
	if seq, _ := reload.FeedState(); seq != 2 {
		t.Fatal("restart lost rollback state")
	}
	if clean, sig, err := reload.Scan(context.Background(), []byte("ABCDEF")); clean || sig != "Other.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
	id, _ := os.ReadFile(filepath.Join(e.DatabaseDir(), "CURRENT"))
	path := filepath.Join(e.DatabaseDir(), "generations", strings.TrimSpace(string(id)), "signatures", "x.ndb")
	_ = os.WriteFile(path, []byte("Tamper:0:*:414243\n"), 0600)
	if err := reload.Reload(); err == nil {
		t.Fatal("tampering accepted")
	}
	if reload.Version() != v {
		t.Fatal("failed reload lost previous snapshot")
	}
}
func TestNativeConcurrentReload(t *testing.T) {
	e := nativeTestEngine(t)
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 20; j++ {
				_, sig, ver, err := e.ScanSnapshot(context.Background(), []byte(testMarker))
				if err != nil || ver == "" || (sig != "Native.Test" && sig != "Updated.Test") {
					t.Errorf("%q %q %v", sig, ver, err)
					return
				}
			}
		}()
	}
	_, err := e.ActivateGeneration(map[string][]byte{"x.ndb": []byte("Updated.Test:0:*:" + hex.EncodeToString([]byte(testMarker)) + "\n")}, 1, manifestHash("update"))
	if err != nil {
		t.Fatal(err)
	}
	wg.Wait()
}
func TestNativeNoSignaturesUnavailable(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	if err := e.EnsureRunning(); err == nil {
		t.Fatal("empty DB healthy")
	}
	if clean, _, err := e.Scan(context.Background(), nil); clean || !errors.Is(err, ErrUnavailable) {
		t.Fatal(clean, err)
	}
}
func FuzzNativeContainers(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte("PK\x03\x04"), []byte("%PDF-"), {0x1f, 0x8b}, []byte("MZ"),
		[]byte("MSCF"), cabFixture(f, 0, [][]byte{[]byte("benign")}, []cabFixtureFile{{"a", 0, 6}}, false),
		{0x78, 0x9f, 0x3e, 0x22}, tnefFixture(tnefAttr(1, 0x0002800c, []byte("benign"))),
		rtfLiteralFixture([]byte("{\\rtf1 benign}")),
		[]byte("MIME-Version: 1.0\r\nContent-Type: text/plain\r\n\r\nx"),
	} {
		f.Add(seed)
	}
	e := nativeTestEngineForFuzz(f)
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<16 {
			t.Skip()
		}
		ctx, cancel := context.WithTimeout(context.Background(), 100000000)
		defer cancel()
		_, _, _ = e.Scan(ctx, data)
	})
}
func nativeTestEngineForFuzz(f *testing.F) *Engine {
	cfg := config.Default()
	cfg.SignatureDBDir = f.TempDir()
	e := New(cfg)
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "x.ndb"), []byte("Test:0:*:414243444546\n"), 0600); err != nil {
		f.Fatal(err)
	}
	if err := e.EnsureRunning(); err != nil {
		f.Fatal(err)
	}
	return e
}

func TestNativeRollbackBeforeInitialization(t *testing.T) {
	e := nativeTestEngine(t)
	files := map[string][]byte{"x.ndb": []byte("Other.Test:0:*:414243444546\n")}
	if _, err := e.ActivateGeneration(files, 8, manifestHash("8")); err != nil {
		t.Fatal(err)
	}
	fresh := New(e.cfg)
	if _, err := fresh.ActivateGeneration(files, 7, manifestHash("7")); err == nil {
		t.Fatal("fresh instance overwrote durable rollback state")
	}
	e.Stop()
	if _, err := e.ActivateGeneration(files, 1, manifestHash("1")); err == nil {
		t.Fatal("stopped instance overwrote durable rollback state")
	}
	_ = os.WriteFile(filepath.Join(e.DatabaseDir(), "CURRENT"), []byte("corrupt"), 0600)
	broken := New(e.cfg)
	if _, err := broken.ActivateGeneration(files, 1, manifestHash("1")); err == nil {
		t.Fatal("corrupt state reset rollback protection")
	}
}

func TestWholeContainerAllowHash(t *testing.T) {
	e := nativeTestEngine(t)
	sample := zipFixture(t, []byte(testMarker))
	h := sha256.Sum256(sample)
	files := map[string][]byte{
		"x.ndb":     []byte("Test:0:*:" + hex.EncodeToString([]byte(testMarker)) + "\n"),
		"allow.sfp": []byte(hex.EncodeToString(h[:]) + ":*:Allowed.Container:73\n"),
	}
	if _, err := e.ActivateGeneration(files, 1, manifestHash("allow")); err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := e.Scan(context.Background(), sample); !clean || sig != "" || err != nil {
		t.Fatal(clean, sig, err)
	}
	delete(files, "x.ndb")
	if _, err := e.ActivateGeneration(files, 2, manifestHash("allowonly")); err == nil {
		t.Fatal("allowlist-only generation accepted as healthy malware database")
	}
}
func TestZIPMembersRejectCountBeforeAllocation(t *testing.T) {
	data := zipFixture(t, []byte("benign"))
	if _, err := zipMembers(data, 0); !errors.Is(err, ErrLimit) {
		t.Fatal(err)
	}
	if m, err := zipMembers(data, 10); err != nil || len(m) != 1 {
		t.Fatal(len(m), err)
	}
	// The reference does not use the end record's directory size.
	data[len(data)-12] = 0
	data[len(data)-11] = 0
	if m, err := zipMembers(data, 10); err != nil || len(m) != 1 {
		t.Fatal(len(m), err)
	}
}

func TestSelfExtractingPEArchive(t *testing.T) {
	e := nativeTestEngine(t)
	sample := append(peValidationFixture(), zipFixture(t, []byte(testMarker))...)
	clean, sig, err := e.Scan(context.Background(), sample)
	if clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
}

func TestMultipleEnginesRespectDurableSequence(t *testing.T) {
	first := nativeTestEngine(t)
	second := New(first.cfg)
	if err := second.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{"x.ndb": []byte("Test:0:*:414243444546\n")}
	if _, err := first.ActivateGeneration(files, 9, manifestHash("9")); err != nil {
		t.Fatal(err)
	}
	if _, err := second.ActivateGeneration(files, 8, manifestHash("8")); err == nil {
		t.Fatal("stale engine replaced a newer durable sequence")
	}
	changed := map[string][]byte{"x.ndb": []byte("Changed:0:*:41424344\n")}
	if _, err := first.ActivateGeneration(changed, 9, manifestHash("9")); err == nil {
		t.Fatal("same manifest sequence accepted different contents")
	}
	release, err := lockUpdate(filepath.Join(first.DatabaseDir(), ".update-lock"))
	if err != nil {
		t.Fatal(err)
	}
	defer release()
	if _, err := second.ActivateGeneration(files, 10, manifestHash("10")); err == nil {
		t.Fatal("concurrent writer bypassed directory lock")
	}
}

func TestInterruptedFirstActivationRetainsLooseDatabase(t *testing.T) {
	e := nativeTestEngine(t)
	if err := os.MkdirAll(filepath.Join(e.DatabaseDir(), "generations", "orphan"), 0700); err != nil {
		t.Fatal(err)
	}
	restarted := New(e.cfg)
	if err := restarted.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := restarted.Scan(context.Background(), []byte(testMarker)); clean || sig != "Native.Test" || err != nil {
		t.Fatal(clean, sig, err)
	}
}

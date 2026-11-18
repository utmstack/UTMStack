package engine

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

func generationCVDFixture(t *testing.T, members map[string][]byte) []byte {
	t.Helper()
	var payload bytes.Buffer
	z := gzip.NewWriter(&payload)
	tw := tar.NewWriter(z)
	names := []string{"first.ndb", "middle.hdb"}
	if _, ok := members["logical.ldb"]; ok {
		names = append(names, "logical.ldb")
	}
	for _, name := range names {
		data := members[name]
		if err := tw.WriteHeader(&tar.Header{Name: name, Size: int64(len(data)), Mode: 0o600}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := z.Close(); err != nil {
		t.Fatal(err)
	}
	header := fmt.Sprintf("ClamAV-VDB:now:123:1:90:%x:unverified:builder:1", md5.Sum(payload.Bytes()))
	return append([]byte(header+strings.Repeat(" ", 512-len(header))), payload.Bytes()...)
}

func TestLooseHashFeedRefreshAndFailedUpdateKeepLastGood(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	pattern := []byte("Pattern.Test:0:*:414243444546\n")
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "rule.ndb"), pattern, 0o600); err != nil {
		t.Fatal(err)
	}
	feed := filepath.Join(cfg.SignatureDBDir, "feed.hdb")
	oldPayload, newPayload := []byte("old malicious fixture"), []byte("new malicious fixture")
	hashRule := func(payload []byte, name string) []byte {
		sum := md5.Sum(payload)
		return []byte(hex.EncodeToString(sum[:]) + ":*:" + name + ":73\n")
	}
	if err := os.WriteFile(feed, hashRule(oldPayload, "Old.Hash"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	if err := e.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	defer e.Stop()
	old := e.current.Load()
	if err := os.WriteFile(feed, hashRule(newPayload, "New.Hash"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}
	current := e.current.Load()
	if current == old || current.db.Version() == old.db.Version() {
		t.Fatal("changed hash source did not publish a new generation")
	}
	if got, err := old.db.Match(context.Background(), oldPayload); err != nil || got != "Old.Hash" {
		t.Fatalf("old snapshot changed: %q, %v", got, err)
	}
	if clean, sig, err := e.Scan(context.Background(), newPayload); clean || sig != "New.Hash" || err != nil {
		t.Fatalf("new hash unavailable: clean=%v signature=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("ABCDEF")); clean || sig != "Pattern.Test" || err != nil {
		t.Fatalf("unchanged pattern unavailable: clean=%v signature=%q err=%v", clean, sig, err)
	}
	if err := os.WriteFile(feed, []byte("invalid hash rule\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err == nil || e.current.Load() != current {
		t.Fatalf("invalid update displaced last-good generation: %v", err)
	}
}

func TestSignedNDBRefreshAndRejectedUpdateKeepLastGood(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	first := map[string][]byte{"feed.ndb": []byte("Old.Rule:0:*:414243444546\n")}
	if _, err := e.ActivateGeneration(first, 1, manifestHash("rule-segment-1")); err != nil {
		t.Fatal(err)
	}
	old := e.current.Load()
	second := map[string][]byte{"feed.ndb": []byte("New.Rule:0:*:4748494a4b4c\n")}
	if _, err := e.ActivateGeneration(second, 2, manifestHash("rule-segment-2")); err != nil {
		t.Fatal(err)
	}
	current := e.current.Load()
	if current == old || current.db.Version() == old.db.Version() {
		t.Fatal("changed rule source did not publish")
	}
	if got, err := old.db.Match(context.Background(), []byte("ABCDEF")); err != nil || got != "Old.Rule" {
		t.Fatalf("prior matcher changed: %q, %v", got, err)
	}
	if got, err := current.db.Match(context.Background(), []byte("GHIJKL")); err != nil || got != "New.Rule" {
		oldGot, _ := current.db.Match(context.Background(), []byte("ABCDEF"))
		full, fullErr := signatures.Compile(second)
		fullGot := ""
		if fullErr == nil {
			fullGot, _ = full.Match(context.Background(), []byte("GHIJKL"))
		}
		t.Fatalf("new database matcher missing rule: new=%q old=%q full=%q fullErr=%v err=%v report=%+v", got, oldGot, fullGot, fullErr, err, current.db.Report())
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("GHIJKL")); clean || sig != "New.Rule" || err != nil {
		t.Fatalf("new rule missing: clean=%v signature=%q error=%v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("ABCDEF")); !clean || sig != "" || err != nil {
		t.Fatalf("old rule survived: clean=%v signature=%q error=%v", clean, sig, err)
	}
	broken := map[string][]byte{"feed.ndb": []byte("Bad.Rule:0:*:zz\n")}
	if _, err := e.ActivateGeneration(broken, 3, manifestHash("rule-segment-3")); err == nil || e.current.Load() != current {
		t.Fatalf("invalid update displaced last-good matcher: %v", err)
	}
}

func TestSignedCVDNDBMemberRefreshKeepsLastGood(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	oldPayload := []byte("old harmless hash payload")
	hashLine := func(data []byte) []byte {
		return []byte(fmt.Sprintf("%x:*:Kept.Hash:73\n", md5.Sum(data)))
	}
	firstMembers := map[string][]byte{
		"first.ndb":  []byte("Old.Rule:0:*:414243444546\n"),
		"middle.hdb": hashLine(oldPayload),
	}
	first := map[string][]byte{"daily.cvd": generationCVDFixture(t, firstMembers)}
	if _, err := e.ActivateGeneration(first, 1, manifestHash("cvd-member-1")); err != nil {
		t.Fatal(err)
	}
	old := e.current.Load()
	secondMembers := map[string][]byte{
		"first.ndb":  []byte("New.Rule:0:*:4748494a4b4c\n"),
		"middle.hdb": hashLine(oldPayload),
	}
	second := map[string][]byte{"daily.cvd": generationCVDFixture(t, secondMembers)}
	if _, err := e.ActivateGeneration(second, 2, manifestHash("cvd-member-2")); err != nil {
		t.Fatal(err)
	}
	current := e.current.Load()
	if current == old || current.db.Version() == old.db.Version() {
		t.Fatal("changed CVD member did not publish")
	}
	if got, err := old.db.Match(context.Background(), []byte("ABCDEF")); err != nil || got != "Old.Rule" {
		t.Fatalf("prior CVD matcher changed: %q, %v", got, err)
	}
	if got, err := current.db.Match(context.Background(), []byte("GHIJKL")); err != nil || got != "New.Rule" {
		t.Fatalf("new CVD member missing: %q, %v", got, err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("ABCDEF")); err != nil || !clean || sig != "" {
		t.Fatalf("old CVD member survived: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if clean, sig, err := e.Scan(context.Background(), oldPayload); err != nil || clean || sig != "Kept.Hash" {
		t.Fatalf("unchanged CVD hash missing: clean=%v sig=%q err=%v", clean, sig, err)
	}
	brokenMembers := map[string][]byte{
		"first.ndb":  []byte("Bad.Rule:0:*:zz\n"),
		"middle.hdb": hashLine(oldPayload),
	}
	broken := map[string][]byte{"daily.cvd": generationCVDFixture(t, brokenMembers)}
	if _, err := e.ActivateGeneration(broken, 3, manifestHash("cvd-member-3")); err == nil || e.current.Load() != current {
		t.Fatalf("invalid CVD update displaced last-good matcher: %v", err)
	}
}

func TestSignedCVDMixedMemberRefreshKeepsLastGood(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	defer e.Stop()
	oldHash, newHash := []byte("old signed hash payload"), []byte("new signed hash payload")
	hashLine := func(data []byte, name string) []byte {
		return []byte(fmt.Sprintf("%x:*:%s:73\n", md5.Sum(data), name))
	}
	first := map[string][]byte{"daily.cvd": generationCVDFixture(t, map[string][]byte{
		"first.ndb": []byte("Old.Rule:0:*:414243444546\n"), "middle.hdb": hashLine(oldHash, "Old.Hash"),
	})}
	if _, err := e.ActivateGeneration(first, 1, manifestHash("cvd-mixed-1")); err != nil {
		t.Fatal(err)
	}
	old := e.current.Load()
	second := map[string][]byte{"daily.cvd": generationCVDFixture(t, map[string][]byte{
		"first.ndb": []byte("New.Rule:0:*:4748494a4b4c\n"), "middle.hdb": hashLine(newHash, "New.Hash"),
	})}
	if _, err := e.ActivateGeneration(second, 2, manifestHash("cvd-mixed-2")); err != nil {
		t.Fatal(err)
	}
	current := e.current.Load()
	if current == old || current.db.Version() == old.db.Version() {
		t.Fatal("mixed CVD update did not publish")
	}
	full, err := signatures.Compile(second)
	if err != nil || current.db.DetectionCount() != full.DetectionCount() ||
		!reflect.DeepEqual(current.db.Report(), full.Report()) {
		t.Fatalf("signed mixed update differs from full compile: %v", err)
	}
	for _, tc := range []struct {
		data     []byte
		old, now string
	}{
		{[]byte("ABCDEF"), "Old.Rule", ""},
		{[]byte("GHIJKL"), "", "New.Rule"},
		{oldHash, "Old.Hash", ""},
		{newHash, "", "New.Hash"},
	} {
		if got, err := old.db.Match(context.Background(), tc.data); err != nil || got != tc.old {
			t.Fatalf("old %q: %q, %v", tc.data, got, err)
		}
		if got, err := current.db.Match(context.Background(), tc.data); err != nil || got != tc.now {
			t.Fatalf("current %q: %q, %v", tc.data, got, err)
		}
	}
	broken := map[string][]byte{"daily.cvd": generationCVDFixture(t, map[string][]byte{
		"first.ndb": []byte("Bad.Rule:0:*:zz\n"), "middle.hdb": hashLine(newHash, "New.Hash"),
	})}
	if _, err := e.ActivateGeneration(broken, 3, manifestHash("cvd-mixed-3")); err == nil || e.current.Load() != current {
		t.Fatalf("invalid mixed update displaced last-good matcher: %v", err)
	}
}

func TestSignedCVDLogicalAndHashMemberRefreshKeepsLastGood(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	defer e.Stop()
	oldHash, newHash := []byte("old logical generation hash"), []byte("new logical generation hash")
	hashLine := func(data []byte, name string) []byte {
		return []byte(fmt.Sprintf("%x:*:%s:73\n", md5.Sum(data), name))
	}
	first := map[string][]byte{"daily.cvd": generationCVDFixture(t, map[string][]byte{
		"first.ndb":   []byte("Old.NDB:0:*:414243444546\n"),
		"middle.hdb":  hashLine(oldHash, "Old.Hash"),
		"logical.ldb": []byte("Old.Logical;Target:0;0;4d4e4f505152\n"),
	})}
	if _, err := e.ActivateGeneration(first, 1, manifestHash("cvd-logical-1")); err != nil {
		t.Fatal(err)
	}
	old := e.current.Load()
	second := map[string][]byte{"daily.cvd": generationCVDFixture(t, map[string][]byte{
		"first.ndb":   []byte("New.NDB:0:*:4748494a4b4c\n"),
		"middle.hdb":  hashLine(newHash, "New.Hash"),
		"logical.ldb": []byte("New.Logical;Target:0;0;535455565758\n"),
	})}
	if _, err := e.ActivateGeneration(second, 2, manifestHash("cvd-logical-2")); err != nil {
		t.Fatal(err)
	}
	current := e.current.Load()
	if current == old || current.db.Version() == old.db.Version() {
		t.Fatal("logical/hash CVD update did not publish")
	}
	full, err := signatures.Compile(second)
	if err != nil || current.db.DetectionCount() != full.DetectionCount() ||
		!reflect.DeepEqual(current.db.Report(), full.Report()) {
		t.Fatalf("signed logical/hash update differs from full compile: %v", err)
	}
	for _, tc := range []struct {
		data     []byte
		old, now string
	}{
		{[]byte("ABCDEF"), "Old.NDB", ""},
		{[]byte("GHIJKL"), "", "New.NDB"},
		{oldHash, "Old.Hash", ""},
		{newHash, "", "New.Hash"},
		{[]byte("MNOPQR"), "Old.Logical", ""},
		{[]byte("STUVWX"), "", "New.Logical"},
	} {
		if got, err := old.db.Match(context.Background(), tc.data); err != nil || got != tc.old {
			t.Fatalf("old %q: %q, %v", tc.data, got, err)
		}
		if got, err := current.db.Match(context.Background(), tc.data); err != nil || got != tc.now {
			t.Fatalf("current %q: %q, %v", tc.data, got, err)
		}
	}
	broken := map[string][]byte{"daily.cvd": generationCVDFixture(t, map[string][]byte{
		"first.ndb":   []byte("New.NDB:0:*:4748494a4b4c\n"),
		"middle.hdb":  hashLine(newHash, "New.Hash"),
		"logical.ldb": []byte("Bad.Logical;Target:0;0;zz\n"),
	})}
	if _, err := e.ActivateGeneration(broken, 3, manifestHash("cvd-logical-3")); err == nil || e.current.Load() != current {
		t.Fatalf("invalid logical update displaced last-good matcher: %v", err)
	}
}

func currentGenerationDir(t *testing.T, e *Engine) string {
	t.Helper()
	id, err := os.ReadFile(filepath.Join(e.DatabaseDir(), "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	return filepath.Join(e.DatabaseDir(), "generations", strings.TrimSpace(string(id)))
}

func TestReloadReusesVerifiedUnchangedGenerationAndObservesLocalChanges(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	feed := map[string][]byte{"feed.ndb": []byte("Feed.Test:0:*:414243444546\n")}
	if _, err := e.ActivateGeneration(feed, 1, manifestHash("reload-reuse")); err != nil {
		t.Fatal(err)
	}
	initial := e.current.Load()
	if err := e.Reload(); err != nil || e.current.Load() != initial {
		t.Fatalf("unchanged verified generation rebuilt: %v", err)
	}
	localPath := filepath.Join(cfg.SignatureDBDir, "local.ndb")
	if err := os.WriteFile(localPath, []byte("Local.Test:0:*:4748494a4b4c\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("GHIJKL")); clean || sig != "Local.Test" || err != nil {
		t.Fatalf("new local rule missing: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if err := os.Remove(localPath); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("GHIJKL")); !clean || sig != "" || err != nil {
		t.Fatalf("removed local rule still matches: clean=%v sig=%q err=%v", clean, sig, err)
	}
	base := e.current.Load()
	if err := e.Reload(); err != nil || e.current.Load() != base {
		t.Fatalf("restored base generation rebuilt: %v", err)
	}
	stored := filepath.Join(currentGenerationDir(t, e), "signatures", "feed.ndb")
	if err := os.WriteFile(stored, []byte("Tampered:0:*:414243444546\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err == nil || e.current.Load() != base {
		t.Fatalf("damaged stored content accepted or last-good lost: %v", err)
	}
}

func TestLooseReloadReusesOnlyByteIdenticalDatabase(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	path := filepath.Join(cfg.SignatureDBDir, "local.ndb")
	if err := os.WriteFile(path, []byte("Alpha:0:*:414141414141\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	if err := e.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	initial := e.current.Load()
	if err := e.Reload(); err != nil || e.current.Load() != initial {
		t.Fatalf("byte-identical loose database rebuilt: %v", err)
	}
	if err := os.WriteFile(path, []byte("Bravo:0:*:424242424242\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err != nil || e.current.Load() == initial {
		t.Fatalf("changed loose database reused: %v", err)
	}
	updated := e.current.Load()
	if clean, sig, err := e.Scan(context.Background(), []byte("BBBBBB")); clean || sig != "Bravo" || err != nil {
		t.Fatalf("changed rule not active: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if err := os.WriteFile(path, []byte("broken\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err == nil || e.current.Load() != updated {
		t.Fatalf("invalid change replaced last-good database: %v", err)
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := e.Reload(); err == nil || e.current.Load() != updated {
		t.Fatalf("empty change replaced last-good database: %v", err)
	}
}

func TestStoredGenerationIdentityDoesNotCompileAndRejectsChangedBytes(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "signatures"), 0o700); err != nil {
		t.Fatal(err)
	}
	// The copy checker intentionally only checks identity. Its caller supplies
	// metadata for incoming bytes that have already passed the full compiler.
	// Even invalid rule text has an identity; it must not be parsed a second time.
	files := map[string][]byte{"test.ndb": []byte("not a valid signature\n")}
	version, err := signatures.ContentVersion(files, databaseLimits)
	if err != nil {
		t.Fatal(err)
	}
	meta := generationMetadata{Sequence: 1, ManifestHash: manifestHash("identity"), ContentVersion: version}
	metadata, err := json.Marshal(meta)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "metadata.json"), metadata, 0o600); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "signatures", "test.ndb")
	if err := os.WriteFile(path, files["test.ndb"], 0o600); err != nil {
		t.Fatal(err)
	}
	if !storedGenerationMatches(dir, meta) {
		t.Fatal("identity-only copy check attempted to compile known content")
	}
	if err := os.WriteFile(path, []byte("changed but equally invalid bytes\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if storedGenerationMatches(dir, meta) {
		t.Fatal("changed stored bytes matched the verified identity")
	}
}

func TestLocalPolicyCannotHideMalformedFeedOnActivation(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	e := New(cfg)
	good := map[string][]byte{"feed.ndb": []byte("Feed.Good:0:*:414243444546\n")}
	if _, err := e.ActivateGeneration(good, 1, manifestHash("first")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "local.ign2"), []byte("Feed.Bad\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	bad := map[string][]byte{"feed.ndb": []byte("Feed.Good:0:*:414243444546\nFeed.Bad:0:*:zz\n")}
	if _, err := e.ActivateGeneration(bad, 2, manifestHash("second")); err == nil {
		t.Fatal("local policy hid a malformed feed rule during activation")
	}
	if sequence, _ := e.FeedState(); sequence != 1 {
		t.Fatalf("failed activation advanced feed state to %d", sequence)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("ABCDEF")); clean || sig != "Feed.Good" || err != nil {
		t.Fatalf("old generation stopped serving: clean=%v sig=%q err=%v", clean, sig, err)
	}
	restarted := New(cfg)
	if err := restarted.EnsureRunning(); err != nil {
		t.Fatal(err)
	}
	if sequence, _ := restarted.FeedState(); sequence != 1 {
		t.Fatalf("failed activation changed durable feed state to %d", sequence)
	}
}

// A stored generation whose rules can no longer be loaded (a file removed by
// another product, disk damage, or a stricter compiler after a binary
// upgrade) must not permanently block a newer, valid signed generation. The
// durable sequence in metadata.json must still enforce anti-rollback.
func TestDamagedStoredGenerationDoesNotBlockNewerUpdate(t *testing.T) {
	for name, damage := range map[string]func(string) error{
		"file removed": func(dir string) error { return os.Remove(filepath.Join(dir, "signatures", "test.ndb")) },
		"no longer compiles": func(dir string) error {
			return os.WriteFile(filepath.Join(dir, "signatures", "test.ndb"), []byte("Gen2.Test:0:*:zz\n"), 0o600)
		},
	} {
		t.Run(name, func(t *testing.T) {
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			first := New(cfg)
			gen2 := map[string][]byte{
				"test.ndb":  []byte("Gen2.Test:0:*:414243444546\n"),
				"extra.ndb": []byte("Gen2.Extra:0:*:4748494a4b4c\n"),
			}
			if _, err := first.ActivateGeneration(gen2, 2, manifestHash("2")); err != nil {
				t.Fatal(err)
			}
			if err := damage(currentGenerationDir(t, first)); err != nil {
				t.Fatal(err)
			}

			// Service restart.
			restarted := New(cfg)
			if err := restarted.EnsureRunning(); err == nil {
				t.Fatal("precondition: damaged generation unexpectedly loaded")
			}

			// Anti-rollback must still hold against the durable sequence 2.
			if _, err := restarted.ActivateGeneration(gen2, 1, manifestHash("1")); err == nil {
				t.Fatal("rollback below durable sequence accepted")
			}
			if _, err := restarted.ActivateGeneration(gen2, 2, manifestHash("other")); err == nil {
				t.Fatal("equivocation at durable sequence accepted")
			}

			// A valid, newer signed generation must be able to repair the endpoint.
			gen3 := map[string][]byte{"test.ndb": []byte("Gen3.Test:0:*:414243444546\n")}
			if _, err := restarted.ActivateGeneration(gen3, 3, manifestHash("3")); err != nil {
				t.Fatalf("newer valid signed generation refused after damage: %v", err)
			}
			if seq, _ := restarted.FeedState(); seq != 3 || restarted.Version() == "" {
				t.Fatalf("engine not serving generation 3: seq=%d version=%q", seq, restarted.Version())
			}
			if clean, sig, err := restarted.Scan(context.Background(), []byte("ABCDEF")); clean || sig != "Gen3.Test" || err != nil {
				t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
			}
			again := New(cfg)
			if err := again.EnsureRunning(); err != nil {
				t.Fatalf("repaired database does not load after restart: %v", err)
			}
		})
	}
}

// Republishing the same signed generation (same sequence, manifest and
// content) must repair a damaged stored copy instead of failing forever. The
// damaged copy is moved aside, not deleted.
func TestSameSignedGenerationRepairsDamagedStoredCopy(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	gen := map[string][]byte{"test.ndb": []byte("Gen.Test:0:*:414243444546\n")}
	version, err := New(cfg).ActivateGeneration(gen, 4, manifestHash("4"))
	if err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	dir := currentGenerationDir(t, e)
	if err := os.WriteFile(filepath.Join(dir, "signatures", "test.ndb"), []byte("Tampered:0:*:414243\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := e.EnsureRunning(); err == nil {
		t.Fatal("precondition: damaged generation unexpectedly loaded")
	}
	changed := map[string][]byte{"test.ndb": []byte("Changed:0:*:414243444546\n")}
	if _, err := e.ActivateGeneration(changed, 4, manifestHash("4")); err == nil {
		t.Fatal("same sequence accepted different content")
	}
	got, err := e.ActivateGeneration(gen, 4, manifestHash("4"))
	if err != nil {
		t.Fatalf("same signed generation could not repair the damaged copy: %v", err)
	}
	if got != version {
		t.Fatalf("repaired version %q, want %q", got, version)
	}
	if err := New(cfg).EnsureRunning(); err != nil {
		t.Fatalf("repaired generation does not load after restart: %v", err)
	}
	entries, err := os.ReadDir(filepath.Join(cfg.SignatureDBDir, "generations"))
	if err != nil {
		t.Fatal(err)
	}
	kept := false
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), ".damaged-") {
			b, err := os.ReadFile(filepath.Join(cfg.SignatureDBDir, "generations", entry.Name(), "signatures", "test.ndb"))
			kept = err == nil && string(b) == "Tampered:0:*:414243\n"
		}
	}
	if !kept {
		t.Fatal("damaged generation was not retained for investigation")
	}
}

// If the verified copy cannot be moved into place (on Windows another
// product may hold a file in it), the damaged copy goes back to its place, so
// CURRENT keeps naming readable rollback state and a later retry repairs it.
func TestInterruptedRepairRestoresDamagedGeneration(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	gen := map[string][]byte{"test.ndb": []byte("Gen.Test:0:*:414243444546\n")}
	if _, err := New(cfg).ActivateGeneration(gen, 4, manifestHash("4")); err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	dir := currentGenerationDir(t, e)
	if err := os.WriteFile(filepath.Join(dir, "signatures", "test.ndb"), []byte("Tampered:0:*:414243\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	renameDirectory = func(from, to string) error {
		if strings.HasPrefix(filepath.Base(from), ".staging-") && to == dir {
			return errors.New("sharing violation")
		}
		return os.Rename(from, to)
	}
	defer func() { renameDirectory = os.Rename }()
	if _, err := e.ActivateGeneration(gen, 4, manifestHash("4")); err == nil {
		t.Fatal("repair reported success although the verified copy was not moved into place")
	}
	if _, err := os.Stat(filepath.Join(dir, "metadata.json")); err != nil {
		t.Fatalf("CURRENT names a generation without rollback state: %v", err)
	}
	if _, err := New(cfg).ActivateGeneration(gen, 1, manifestHash("1")); err == nil || !strings.Contains(err.Error(), "rollback") {
		t.Fatalf("rollback state lost after an interrupted repair: %v", err)
	}
	renameDirectory = os.Rename
	if _, err := New(cfg).ActivateGeneration(gen, 4, manifestHash("4")); err != nil {
		t.Fatalf("retry did not repair the damaged copy: %v", err)
	}
	if err := New(cfg).EnsureRunning(); err != nil {
		t.Fatalf("repaired generation does not load: %v", err)
	}
}

// Without trustworthy rollback state (missing metadata), updates fail closed.
func TestMissingGenerationMetadataFailsClosed(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	gen := map[string][]byte{"test.ndb": []byte("Gen.Test:0:*:414243444546\n")}
	e := New(cfg)
	if _, err := e.ActivateGeneration(gen, 5, manifestHash("5")); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(currentGenerationDir(t, e), "metadata.json")); err != nil {
		t.Fatal(err)
	}
	restarted := New(cfg)
	if _, err := restarted.ActivateGeneration(gen, 1, manifestHash("1")); err == nil {
		t.Fatal("missing rollback metadata reset rollback protection")
	}
}

// After a remote generation is active, administrator files placed in
// signature_db_dir (the documented local workflow) must load together with
// the generation, including after a service restart or in engine-check.
func TestLooseFilesLoadAlongsideActivatedGeneration(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	remote := []byte("Remote.Test:0:*:" + hex.EncodeToString([]byte("REMOTE-FEED-MARKER")) + "\n")
	running := New(cfg)
	if _, err := running.ActivateGeneration(map[string][]byte{"remote.ndb": remote}, 1, manifestHash("gen-1")); err != nil {
		t.Fatal(err)
	}
	const custom = "INCIDENT-CUSTOM-MARKER"
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "custom.ndb"),
		[]byte("Incident.Custom:0:*:"+hex.EncodeToString([]byte(custom))+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// Service restart / engine-check.
	e := New(cfg)
	if err := e.EnsureRunning(); err != nil {
		t.Fatalf("valid administrator rule made the database fail: %v", err)
	}
	for data, want := range map[string]string{"prefix " + custom + " suffix": "Incident.Custom", "REMOTE-FEED-MARKER": "Remote.Test"} {
		clean, sig, err := e.Scan(context.Background(), []byte(data))
		if clean || sig != want || err != nil {
			t.Fatalf("%q: clean=%v sig=%q err=%v (signatures=%d)", data, clean, sig, err, e.SignatureCount())
		}
	}
	if seq, _ := e.FeedState(); seq != 1 {
		t.Fatalf("feed state lost: %d", seq)
	}
	// The local content is part of the database identity, so clean verdicts
	// cached without it are withdrawn.
	if e.Version() == running.Version() {
		t.Fatal("database identity ignores local signature files")
	}
	// A later signed activation keeps the local files in the running engine.
	next := []byte("Remote.Next:0:*:" + hex.EncodeToString([]byte("NEXT-FEED-MARKER")) + "\n")
	version, err := e.ActivateGeneration(map[string][]byte{"remote.ndb": next}, 2, manifestHash("gen-2"))
	if err != nil {
		t.Fatal(err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte(custom)); clean || sig != "Incident.Custom" || err != nil {
		t.Fatalf("activation dropped local rule: clean=%v sig=%q err=%v", clean, sig, err)
	}
	restarted := New(cfg)
	if err := restarted.EnsureRunning(); err != nil || restarted.Version() != version {
		t.Fatalf("restart identity %q, activation identity %q, err=%v", restarted.Version(), version, err)
	}
}

// An unsupported rule dropped into signature_db_dir must not be silently
// accepted as a healthy database once a remote generation exists. The signed
// generation keeps serving, and a signed update is not blocked by it.
func TestUnsupportedLooseRuleReportedAlongsideGeneration(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	remote := []byte("Remote.Test:0:*:" + hex.EncodeToString([]byte("REMOTE-FEED-MARKER")) + "\n")
	if _, err := New(cfg).ActivateGeneration(map[string][]byte{"remote.ndb": remote}, 1, manifestHash("gen-1")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "custom.ndb"), []byte("Bad.Rule:0:*:zz\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	err := e.EnsureRunning()
	if !errors.Is(err, ErrLocalSignatures) {
		t.Fatalf("unsupported rule in signature_db_dir not reported: %v", err)
	}
	if !strings.Contains(err.Error(), "custom.ndb") {
		t.Fatalf("report does not name the local file: %v", err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("REMOTE-FEED-MARKER")); clean || sig != "Remote.Test" || err != nil {
		t.Fatalf("signed generation stopped serving: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if seq, _ := e.FeedState(); seq != 1 {
		t.Fatalf("feed state lost: %d", seq)
	}
	next := []byte("Remote.Next:0:*:" + hex.EncodeToString([]byte("NEXT-FEED-MARKER")) + "\n")
	if _, err := e.ActivateGeneration(map[string][]byte{"remote.ndb": next}, 2, manifestHash("gen-2")); err != nil {
		t.Fatalf("local file blocked a signed update: %v", err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("NEXT-FEED-MARKER")); clean || sig != "Remote.Next" || err != nil {
		t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
	}
	if err := New(cfg).EnsureRunning(); !errors.Is(err, ErrLocalSignatures) {
		t.Fatalf("restart does not report the local file: %v", err)
	}
}

// A local file with the same name as a feed artifact would make the combined
// database ambiguous; it is reported instead of shadowing either one.
func TestLooseFileNameCollisionReported(t *testing.T) {
	cfg := config.Default()
	cfg.SignatureDBDir = t.TempDir()
	remote := []byte("Remote.Test:0:*:414243444546\n")
	if _, err := New(cfg).ActivateGeneration(map[string][]byte{"daily.ndb": remote}, 1, manifestHash("gen-1")); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfg.SignatureDBDir, "daily.ndb"), []byte("Local.Test:0:*:4748494a4b4c\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	e := New(cfg)
	if err := e.EnsureRunning(); !errors.Is(err, ErrLocalSignatures) {
		t.Fatalf("name collision not reported: %v", err)
	}
	if clean, sig, err := e.Scan(context.Background(), []byte("ABCDEF")); clean || sig != "Remote.Test" || err != nil {
		t.Fatalf("clean=%v sig=%q err=%v", clean, sig, err)
	}
}

package feed

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

type preparedEngineSpy struct {
	*engine.Engine
	preparedCalls int
	preparedFiles int
}

func (s *preparedEngineSpy) ActivatePreparedGeneration(files map[string][]byte, prepared map[string]*signatures.PreparedHashList, sequence uint64, hash string) (string, error) {
	s.preparedCalls++
	s.preparedFiles = len(prepared)
	return s.Engine.ActivatePreparedGeneration(files, prepared, sequence, hash)
}

func TestSignedHashListUsesNativeInputAndPreservesRestartAndRepair(t *testing.T) {
	fix := newFeedFixture(t)
	cfg := fix.f.cfg
	cfg.SignatureDBDir = t.TempDir()
	spy := &preparedEngineSpy{Engine: engine.New(cfg)}
	fix.f.engine = spy
	payload := []byte("an inert signed hash-list fixture body")
	digest := sha256.Sum256(payload)
	list := []byte(hex.EncodeToString(digest[:]) + "\n")
	checksum := sha256.Sum256(list)
	fix.manifest.Files = []ManifestFile{{Name: "intel.hsb", URL: "intel.list", Format: "sha256-list", SHA256: hex.EncodeToString(checksum[:])}}
	fix.resources["/signatures/intel.list"] = list
	fix.sign(t)
	version, err := fix.f.updateOnce()
	if err != nil {
		t.Fatal(err)
	}
	if spy.preparedCalls != 1 || spy.preparedFiles != 1 {
		t.Fatalf("native hash-list path not used: calls=%d files=%d", spy.preparedCalls, spy.preparedFiles)
	}
	current, err := os.ReadFile(filepath.Join(cfg.SignatureDBDir, "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(cfg.SignatureDBDir, "generations", string(bytes.TrimSpace(current)), "signatures", "intel.hsb")
	stored, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := hex.EncodeToString(digest[:]) + ":*:UTMStack.Intel.SHA256." + hex.EncodeToString(digest[:]) + ":73\n"
	if string(stored) != want {
		t.Fatalf("durable compatibility changed: %q", stored)
	}
	if _, err := fix.f.updateOnce(); err != nil || spy.preparedCalls != 1 {
		t.Fatalf("unchanged generation was activated again: calls=%d err=%v", spy.preparedCalls, err)
	}
	restarted := engine.New(cfg)
	if err := restarted.EnsureRunning(); err != nil || restarted.Version() != version {
		t.Fatalf("durable text restart differs: version=%q err=%v", restarted.Version(), err)
	}
	if clean, sig, err := restarted.Scan(context.Background(), payload); clean || sig != "UTMStack.Intel.SHA256."+hex.EncodeToString(digest[:]) || err != nil {
		t.Fatalf("restart lost native indicators: clean=%v sig=%q err=%v", clean, sig, err)
	}
	if err := os.WriteFile(path, []byte("damaged stored source\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	repair := &preparedEngineSpy{Engine: engine.New(cfg)}
	if err := repair.EnsureRunning(); err == nil {
		t.Fatal("damaged copy unexpectedly loaded")
	}
	fix.f.engine = repair
	if repaired, err := fix.f.updateOnce(); err != nil || repaired != version || repair.preparedCalls != 1 {
		t.Fatalf("same-sequence native repair failed: version=%q calls=%d err=%v", repaired, repair.preparedCalls, err)
	}
	if stored, err := os.ReadFile(path); err != nil || string(stored) != want {
		t.Fatalf("same-sequence repair changed durable identity: %q err=%v", stored, err)
	}
	if seq, _ := repair.FeedState(); seq != fix.manifest.Sequence {
		t.Fatalf("repair sequence=%d", seq)
	}
}

func TestSignedGenerationKeepsReplayStateAcrossEngineRestart(t *testing.T) {
	fix := newFeedFixture(t)
	cfg := fix.f.cfg
	cfg.SignatureDBDir = t.TempDir()
	first := engine.New(cfg)
	fix.f.engine = first
	version, err := fix.f.updateOnce()
	if err != nil {
		t.Fatal(err)
	}
	if first.SignatureCount() != 1 {
		t.Fatal("real signature compiler did not load fixture")
	}
	second := engine.New(cfg)
	if err := second.Reload(); err != nil {
		t.Fatal(err)
	}
	if second.Version() != version {
		t.Fatal("persisted content version changed")
	}
	fix.f.engine = second
	fix.manifest.Sequence = 1
	fix.sign(t)
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("restart allowed signed rollback")
	}
	fix.manifest.Sequence = 3
	unsupported := []byte("unsupported bytecode fixture")
	hash := sha256.Sum256(unsupported)
	fix.manifest.Files = []ManifestFile{{Name: "unsupported.cbc", URL: "unsupported.cbc", SHA256: hex.EncodeToString(hash[:])}}
	fix.resources["/signatures/unsupported.cbc"] = unsupported
	fix.sign(t)
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("unsupported signed generation silently loaded")
	}
	if seq, _ := second.FeedState(); seq != 2 || second.Version() != version {
		t.Fatal("rejected rules changed active sequence or content")
	}
	third := engine.New(cfg)
	if err := third.Reload(); err != nil {
		t.Fatal(err)
	}
	if seq, _ := third.FeedState(); seq != 2 || third.Version() != version {
		t.Fatal("rejected rules changed durable generation")
	}
}

// An update cycle whose signed manifest is the active generation must not
// download, compile and durably rewrite the whole generation again.
func TestUnchangedSignedGenerationIsNotRedownloaded(t *testing.T) {
	fix := newFeedFixture(t)
	cfg := fix.f.cfg
	cfg.SignatureDBDir = t.TempDir()
	fix.f.engine = engine.New(cfg)
	first, err := fix.f.updateOnce()
	if err != nil {
		t.Fatal(err)
	}
	if got := fix.artifactHits.Load(); got != 1 {
		t.Fatalf("precondition: first cycle downloaded %d artifacts", got)
	}
	current, err := os.ReadFile(filepath.Join(cfg.SignatureDBDir, "CURRENT"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := fix.f.updateOnce()
	if err != nil {
		t.Fatalf("identical manifest must remain a successful no-op: %v", err)
	}
	if second != first {
		t.Fatalf("version changed on no-op cycle: %q -> %q", first, second)
	}
	if got := fix.artifactHits.Load(); got != 1 {
		t.Fatalf("unchanged generation re-downloaded: %d artifact requests after two cycles", got)
	}
	if again, err := os.ReadFile(filepath.Join(cfg.SignatureDBDir, "CURRENT")); err != nil || !bytes.Equal(again, current) {
		t.Fatalf("no-op cycle changed the active generation pointer: %v", err)
	}
}

// Hash-list indicators for the empty object never fire on empty files or empty
// archive members, while the other indicators of the same list still detect.
func TestEmptyObjectDigestDoesNotDetectEmptyFiles(t *testing.T) {
	malicious := []byte("an ordinary indicator-listed payload body")
	md5Sum, sha1Sum, sha256Sum := md5.Sum(malicious), sha1.Sum(malicious), sha256.Sum256(malicious)
	for _, tc := range []struct {
		format, name, empty, listed, algorithm string
	}{
		{"md5-list", "intel.hdb", "d41d8cd98f00b204e9800998ecf8427e", hex.EncodeToString(md5Sum[:]), "MD5"},
		{"sha1-list", "intel.hsb", "da39a3ee5e6b4b0d3255bfef95601890afd80709", hex.EncodeToString(sha1Sum[:]), "SHA1"},
		{"sha256-list", "intel.hsb", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", hex.EncodeToString(sha256Sum[:]), "SHA256"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			list := []byte(tc.empty + "\n" + tc.listed + "\n")
			converted, err := NormalizeArtifact(ManifestFile{Name: tc.name, Format: tc.format}, list)
			if err != nil {
				t.Fatal(err)
			}
			cfg := config.Default()
			cfg.SignatureDBDir = t.TempDir()
			e := engine.New(cfg)
			if _, err := e.ActivateGeneration(map[string][]byte{tc.name: converted}, 1, "0000000000000000000000000000000000000000000000000000000000000001"); err != nil {
				t.Fatal(err)
			}
			if clean, sig, err := e.Scan(context.Background(), []byte{}); sig != "" {
				t.Errorf("empty file detected: clean=%v sig=%q err=%v", clean, sig, err)
			}
			var b bytes.Buffer
			w := zip.NewWriter(&b)
			if _, err := w.Create("placeholder.txt"); err != nil {
				t.Fatal(err)
			}
			if f, err := w.Create("readme.txt"); err != nil {
				t.Fatal(err)
			} else if _, err := f.Write([]byte("an ordinary benign document body")); err != nil {
				t.Fatal(err)
			}
			if err := w.Close(); err != nil {
				t.Fatal(err)
			}
			if clean, sig, err := e.Scan(context.Background(), b.Bytes()); sig != "" {
				t.Errorf("archive with an empty member detected: clean=%v sig=%q err=%v", clean, sig, err)
			}
			want := "UTMStack.Intel." + tc.algorithm + "." + tc.listed
			if clean, sig, err := e.Scan(context.Background(), malicious); clean || sig != want || err != nil {
				t.Errorf("listed indicator lost: clean=%v sig=%q err=%v", clean, sig, err)
			}
		})
	}
}

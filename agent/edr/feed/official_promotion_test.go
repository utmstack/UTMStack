package feed

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/cache"
	"github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/engine"
	"github.com/utmstack/UTMStack/agent/edr/signatures"
)

// Run with EDR_OFFICIAL_OLD_DB and EDR_OFFICIAL_NEW_DAILY pointing to the
// pinned publisher-signed CVDs. The ordinary test suite skips this large,
// externally supplied integration fixture.
func TestOfficialCVDAndThreatWindSignedPromotion(t *testing.T) {
	oldDir, nextDaily := os.Getenv("EDR_OFFICIAL_OLD_DB"), os.Getenv("EDR_OFFICIAL_NEW_DAILY")
	if oldDir == "" || nextDaily == "" {
		t.Skip("set EDR_OFFICIAL_OLD_DB and EDR_OFFICIAL_NEW_DAILY")
	}
	read := func(path string) []byte {
		t.Helper()
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	mainPath := filepath.Join(oldDir, "main.cvd")
	oldDailyPath := filepath.Join(oldDir, "daily.cvd")
	bytecodePath := filepath.Join(oldDir, "bytecode.cvd")
	oldDaily := read(oldDailyPath)
	newDaily := read(nextDaily)
	delta := officialDeltaNDBFixture(t, oldDaily, newDaily)
	oldDaily, newDaily = nil, nil
	intelPayload := []byte("inert ThreatWind signed-hash-list promotion control")
	intelHash := sha256.Sum256(intelPayload)
	intelList := []byte(hex.EncodeToString(intelHash[:]) + "\n")
	intelName := "UTMStack.Intel.SHA256." + hex.EncodeToString(intelHash[:])

	fix := newFeedFixture(t)
	cfg := fix.f.cfg
	cfg.SignatureDBDir = t.TempDir()
	cfg.ScanConcurrency = 1
	cfg.ScanLimits = config.ScanLimits{MaxFileSizeMB: 200, MaxScanSizeMB: 400}
	fix.f.cfg = cfg
	eng := engine.New(cfg)
	defer eng.Stop()
	fix.f.engine = eng
	verdictCache, err := cache.Open(filepath.Join(t.TempDir(), "edr.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer verdictCache.Close()
	fix.f.cache = verdictCache
	// The mirror runs in this test process. Stream the large CVDs from disk
	// rather than retaining both snapshots in the fake server's heap; this
	// makes the capped client-memory measurement closer to a remote mirror.
	var servedDaily atomic.Value
	servedDaily.Store(oldDailyPath)
	fix.handlers = map[string]http.HandlerFunc{
		"/signatures/main.cvd": func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, mainPath) },
		"/signatures/daily.cvd": func(w http.ResponseWriter, r *http.Request) {
			http.ServeFile(w, r, servedDaily.Load().(string))
		},
		"/signatures/bytecode.cvd": func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, bytecodePath) },
	}
	resource := func(name, url, format string, data []byte) ManifestFile {
		t.Helper()
		fix.resources["/signatures/"+url] = data
		digest := sha256.Sum256(data)
		return ManifestFile{Name: name, URL: url, Format: format, SHA256: hex.EncodeToString(digest[:])}
	}
	fileResource := func(name, url, path string) ManifestFile {
		t.Helper()
		file, err := os.Open(path)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		hash := sha256.New()
		if _, err := io.Copy(hash, file); err != nil {
			t.Fatal(err)
		}
		return ManifestFile{Name: name, URL: url, SHA256: hex.EncodeToString(hash.Sum(nil))}
	}
	mainFile := fileResource("main.cvd", "main.cvd", mainPath)
	oldDailyFile := fileResource("daily.cvd", "daily.cvd", oldDailyPath)
	bytecodeFile := fileResource("bytecode.cvd", "bytecode.cvd", bytecodePath)
	intelFile := resource("threatwind.hsb", "tw.list", "sha256-list", intelList)
	fix.manifest.Files = []ManifestFile{mainFile, oldDailyFile, bytecodeFile, intelFile}
	fix.sign(t)
	oldVersion, err := fix.f.updateOnce()
	if err != nil {
		t.Fatalf("old signed generation: %v", err)
	}
	if got := eng.SignatureCount(); got != 3628165 {
		t.Fatalf("old signature count = %d", got)
	}
	assertOfficialPromotionVerdict(t, eng, []byte(`X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`), "Eicar-Test-Signature")
	assertOfficialPromotionVerdict(t, eng, intelPayload, intelName)
	if clean, sig, err := eng.Scan(context.Background(), delta); err != nil || !clean || sig != "" {
		t.Fatalf("old feed delta verdict: clean=%v sig=%q err=%v", clean, sig, err)
	}
	assertOfficialContainers(t, eng.LoadReport(), "28158")
	deltaHash := sha256.Sum256(delta)
	if err := verdictCache.Store(cache.VerdictRecord{SHA256: hex.EncodeToString(deltaHash[:]), Verdict: cache.VerdictClean, SigDBVersion: oldVersion}); err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var wg sync.WaitGroup
	var mu sync.Mutex
	generations := map[string]int{}
	scanErrors := 0
	wg.Add(1)
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(10 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
			clean, sig, version, err := eng.ScanSnapshot(context.Background(), []byte(`X5O!P%@AP[4\PZX54(P^)7CC)7}$EICAR-STANDARD-ANTIVIRUS-TEST-FILE!$H+H*`))
			mu.Lock()
			generations[version]++
			if err != nil || clean || sig != "Eicar-Test-Signature" {
				scanErrors++
			}
			mu.Unlock()
		}
	}()
	newDailyFile := fileResource("daily.cvd", "daily.cvd", nextDaily)
	servedDaily.Store(nextDaily)
	fix.manifest.Sequence, fix.manifest.Version = 3, "official-daily-28167"
	fix.manifest.Files = []ManifestFile{mainFile, newDailyFile, bytecodeFile, intelFile}
	fix.sign(t)
	start := time.Now()
	newVersion, err := fix.f.updateOnce()
	reloadSeconds := time.Since(start).Seconds()
	if err == nil {
		time.Sleep(150 * time.Millisecond)
	}
	cancel()
	wg.Wait()
	if err != nil {
		t.Fatalf("new signed generation: %v", err)
	}
	if oldVersion == newVersion || eng.SignatureCount() != 3628229 {
		t.Fatalf("new version/count = %q/%d", newVersion, eng.SignatureCount())
	}
	assertOfficialContainers(t, eng.LoadReport(), "28167")
	assertOfficialPromotionVerdict(t, eng, delta, "Unix.Trojan.Generic-10060706-0")
	assertOfficialPromotionVerdict(t, eng, intelPayload, intelName)
	if record, found, err := verdictCache.Lookup(hex.EncodeToString(deltaHash[:])); err != nil || !found || !record.Stale {
		t.Fatalf("old clean verdict not invalidated: found=%v record=%+v err=%v", found, record, err)
	}
	mu.Lock()
	oldScans, newScans, errorsCount := generations[oldVersion], generations[newVersion], scanErrors
	unexpected := len(generations) != 2
	mu.Unlock()
	if oldScans == 0 || newScans == 0 || errorsCount != 0 || unexpected {
		t.Fatalf("concurrent scan generations=%v errors=%d", generations, errorsCount)
	}
	activeSequence, activeHash := eng.FeedState()
	if activeSequence != 3 || activeHash == "" {
		t.Fatalf("active sequence/hash = %d/%q", activeSequence, activeHash)
	}
	bad := fix.manifest.Files
	bad[1].SHA256 = strings.Repeat("0", 64)
	fix.manifest.Sequence, fix.manifest.Version, fix.manifest.Files = 4, "bad-daily", bad
	fix.sign(t)
	if _, err := fix.f.updateOnce(); err == nil {
		t.Fatal("bad artifact digest was accepted")
	}
	if seq, hash := eng.FeedState(); seq != activeSequence || hash != activeHash || eng.Version() != newVersion {
		t.Fatalf("failed update changed active state: %d/%q version=%q", seq, hash, eng.Version())
	}
	assertOfficialPromotionVerdict(t, eng, delta, "Unix.Trojan.Generic-10060706-0")
	assertOfficialPromotionVerdict(t, eng, intelPayload, intelName)
	t.Logf("official signed promotion: old=%d new=%d scans=%d/%d scan_errors=%d reload_seconds=%.3f rollback=true threatwind=true stale_cache=true", 3628165, 3628229, oldScans, newScans, errorsCount, reloadSeconds)
}

func assertOfficialPromotionVerdict(t *testing.T, eng *engine.Engine, data []byte, want string) {
	t.Helper()
	clean, sig, err := eng.Scan(context.Background(), data)
	if err != nil || clean || sig != want {
		t.Fatalf("verdict: clean=%v sig=%q want=%q err=%v", clean, sig, want, err)
	}
}

func assertOfficialContainers(t *testing.T, report signatures.Report, daily string) {
	t.Helper()
	if len(report.Containers) != 3 {
		t.Fatalf("container count = %d", len(report.Containers))
	}
	seenDaily := false
	for _, c := range report.Containers {
		if !c.IntegrityVerified || !c.Authenticated {
			t.Fatalf("container not authenticated: %+v", c)
		}
		if c.File == "daily.cvd" {
			seenDaily = c.Version == daily
		}
	}
	if !seenDaily {
		t.Fatalf("daily version %q missing", daily)
	}
}

func officialDeltaNDBFixture(t *testing.T, oldDaily, newDaily []byte) []byte {
	t.Helper()
	getRule := func(data []byte) string {
		t.Helper()
		if len(data) <= 512 {
			t.Fatal("short daily CVD")
		}
		gz, err := gzip.NewReader(bytes.NewReader(data[512:]))
		if err != nil {
			t.Fatal(err)
		}
		defer gz.Close()
		tr := tar.NewReader(gz)
		for {
			h, err := tr.Next()
			if err != nil {
				t.Fatal(err)
			}
			if h.Name != "daily.ndb" {
				continue
			}
			if h.Size > 1<<20 {
				t.Fatal("daily.ndb exceeds fixture limit")
			}
			body, err := io.ReadAll(tr)
			if err != nil {
				t.Fatal(err)
			}
			for _, line := range strings.Split(string(body), "\n") {
				if strings.HasPrefix(line, "Unix.Trojan.Generic-10060706-0:") {
					return line
				}
			}
			return ""
		}
	}
	if getRule(oldDaily) != "" {
		t.Fatal("new NDB rule unexpectedly exists in the old CVD")
	}
	fields := strings.Split(getRule(newDaily), ":")
	if len(fields) != 4 || fields[1] != "0" || fields[2] != "*" || len(fields[3]) != 192 {
		t.Fatalf("new NDB rule shape changed: %v", fields)
	}
	pattern, err := hex.DecodeString(fields[3])
	if err != nil {
		t.Fatal(err)
	}
	return append(append([]byte("INERT-SIGNATURE-PROBE:"), pattern...), []byte(":END")...)
}

package dependency

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/buildinfo"
	"github.com/utmstack/UTMStack/agent/edr/release"
)

func pairedReleaseUpdate(t *testing.T, old bool) (*releaseEDRUpdate, *[]string) {
	t.Helper()
	u := newReleaseEDRUpdate(t.TempDir())
	u.identity = release.Identity{PackageVersion: EDRVersion, Implementation: buildinfo.ImplementationVersion, OS: "windows", Architecture: "amd64", ServiceName: "utmstack_edr_windows_amd64.exe", ProviderName: "utmstack_amsi.dll"}
	u.names = []string{u.identity.ServiceName, u.identity.ProviderName, u.identity.ServiceName + ".release.json"}
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{17}, ed25519.SeedSize))
	u.key = func() (ed25519.PublicKey, error) { return private.Public().(ed25519.PublicKey), nil }
	source := t.TempDir()
	for name, body := range map[string]string{u.identity.ServiceName: "new service", u.identity.ProviderName: "new provider"} {
		if err := os.WriteFile(filepath.Join(source, name), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := release.Pack(u.archive, source, private, u.identity); err != nil {
		t.Fatal(err)
	}
	if old {
		for name, body := range map[string]string{u.identity.ServiceName: "old service", u.identity.ProviderName: "old provider", u.identity.ServiceName + ".release.json": "old manifest"} {
			if err := os.WriteFile(filepath.Join(u.base, name), []byte(body), 0600); err != nil {
				t.Fatal(err)
			}
		}
	}
	var calls []string
	running := old
	u.version = func(path string) error {
		calls = append(calls, "version")
		b, err := os.ReadFile(path)
		if err != nil || string(b) != "new service" {
			return errors.New("wrong staged service")
		}
		return nil
	}
	u.status = func() (edrServiceState, error) {
		calls = append(calls, "status")
		return edrServiceState{installed: old, running: running}, nil
	}
	u.stop = func() error { calls = append(calls, "stop"); running = false; return nil }
	u.start = func() error { calls = append(calls, "start"); running = true; return nil }
	u.ready = func(time.Time) error { calls = append(calls, "ready"); return nil }
	u.install = func() error { calls = append(calls, "install"); return nil }
	return u, &calls
}

func TestSignedEDRPairPromotionAndServiceRestart(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{u.identity.ServiceName: "new service", u.identity.ProviderName: "new provider"} {
		body, err := os.ReadFile(filepath.Join(u.base, name))
		if err != nil || string(body) != want {
			t.Fatalf("%s = %q, %v", name, body, err)
		}
	}
	if err := u.verifyInstalled(); err != nil {
		t.Fatalf("promoted signed pair failed installed verification: %v", err)
	}
	if u.backupDir != "" || u.stageDir != "" || strings.Join(*calls, ",") != "version,status,stop,start,status,ready,status" {
		t.Fatalf("unexpected transaction state or service calls: %q, backup=%q stage=%q", *calls, u.backupDir, u.stageDir)
	}
}

func TestSignedEDRPairRestartFailureRestoresAllOldFiles(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	starts := 0
	u.start = func() error {
		*calls = append(*calls, "start")
		starts++
		if starts == 1 {
			return errors.New("replacement start failed")
		}
		return nil
	}
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err == nil || !strings.Contains(err.Error(), "replacement start failed") {
		t.Fatalf("replacement failure not reported: %v", err)
	}
	for name, want := range map[string]string{u.identity.ServiceName: "old service", u.identity.ProviderName: "old provider", u.identity.ServiceName + ".release.json": "old manifest"} {
		body, err := os.ReadFile(filepath.Join(u.base, name))
		if err != nil || string(body) != want {
			t.Fatalf("%s was not restored: %q, %v", name, body, err)
		}
	}
	if starts != 2 || u.backupDir != "" {
		t.Fatalf("previous service not recovered: starts=%d backup=%q calls=%q", starts, u.backupDir, *calls)
	}
}

func TestSignedEDRPairUnreadyReplacementRestoresOldFiles(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	u.ready = func(time.Time) error {
		*calls = append(*calls, "ready")
		return errors.New("AMSI local pipe did not start")
	}
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err == nil || !strings.Contains(err.Error(), "AMSI local pipe did not start") {
		t.Fatalf("unready release was committed: %v", err)
	}
	for name, want := range map[string]string{u.identity.ServiceName: "old service", u.identity.ProviderName: "old provider", u.identity.ServiceName + ".release.json": "old manifest"} {
		body, err := os.ReadFile(filepath.Join(u.base, name))
		if err != nil || string(body) != want {
			t.Fatalf("%s was not restored: %q, %v", name, body, err)
		}
	}
	if strings.Join(*calls, ",") != "version,status,stop,start,status,ready,status,stop,start" {
		t.Fatalf("replacement rollback calls: %q", *calls)
	}
}

func TestSignedEDRPairExitAfterReadyRestoresOldFiles(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	running := true
	installed := true
	u.status = func() (edrServiceState, error) {
		*calls = append(*calls, "status")
		return edrServiceState{installed: installed, running: running}, nil
	}
	u.stop = func() error { *calls = append(*calls, "stop"); running = false; return nil }
	u.start = func() error { *calls = append(*calls, "start"); running = true; return nil }
	u.ready = func(time.Time) error { *calls = append(*calls, "ready"); running = false; return nil }
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err == nil || !strings.Contains(err.Error(), "stopped before commit") {
		t.Fatalf("exited replacement was committed: %v", err)
	}
	if !running || !installed || u.backupDir != "" {
		t.Fatalf("previous service not restored: running=%v installed=%v calls=%q", running, installed, *calls)
	}
}

func TestEDRReleaseReadyRequiresFreshHealthyProcessAndAMSI(t *testing.T) {
	u := newReleaseEDRUpdate(t.TempDir())
	u.expectedServiceSHA = strings.Repeat("a", 64)
	start := time.Now().UTC()
	status := edrRuntimeStatus{
		InstanceID: "new-process", ProcessID: 1234,
		StartedAt: start.Add(time.Second).Format(time.RFC3339Nano),
		Running:   true, EngineHealthy: true,
		EngineImplementation: u.identity.Implementation,
		ServiceSHA256:        u.expectedServiceSHA,
		SensorAMSI:           true, AMSILocalReady: true,
	}
	write := func() {
		t.Helper()
		data, err := json.Marshal(status)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(u.base, "status.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	write()
	if err := u.checkReady(start); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		mut  func(*edrRuntimeStatus)
	}{
		{"stale", func(s *edrRuntimeStatus) { s.StartedAt = start.Add(-time.Second).Format(time.RFC3339Nano) }},
		{"no process identity", func(s *edrRuntimeStatus) { s.InstanceID = "" }},
		{"wrong release", func(s *edrRuntimeStatus) { s.EngineImplementation = "different" }},
		{"wrong executable", func(s *edrRuntimeStatus) { s.ServiceSHA256 = strings.Repeat("b", 64) }},
		{"unhealthy", func(s *edrRuntimeStatus) { s.EngineHealthy = false }},
		{"AMSI unavailable", func(s *edrRuntimeStatus) { s.AMSILocalReady = false }},
		{"not running", func(s *edrRuntimeStatus) { s.Running = false }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bad := status
			tc.mut(&bad)
			data, _ := json.Marshal(bad)
			if err := os.WriteFile(filepath.Join(u.base, "status.json"), data, 0600); err != nil {
				t.Fatal(err)
			}
			if err := u.checkReady(start); err == nil {
				t.Fatal("unready status accepted")
			}
		})
	}
	write()
	if err := os.Remove(filepath.Join(u.base, "status.json")); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(u.archive, filepath.Join(u.base, "status.json")); err == nil {
		if err := u.checkReady(start); err == nil {
			t.Fatal("symlinked status accepted")
		}
	}
}

func TestSignedEDRRejectsUntrustedArchiveBeforeStopOrVersionExecution(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	u.key = func() (ed25519.PublicKey, error) { return nil, errors.New("no pinned publisher key") }
	if err := u.promote(); err == nil {
		t.Fatal("untrusted release accepted")
	}
	if len(*calls) != 0 {
		t.Fatalf("untrusted release touched service or executed bytes: %q", *calls)
	}
	body, err := os.ReadFile(filepath.Join(u.base, u.identity.ServiceName))
	if err != nil || string(body) != "old service" {
		t.Fatalf("old service changed: %q, %v", body, err)
	}
}

func TestSignedEDRRejectsChangedProviderBeforeStop(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	data, err := os.ReadFile(u.archive)
	if err != nil {
		t.Fatal(err)
	}
	at := bytes.Index(data, []byte("new provider"))
	if at < 0 {
		t.Fatal("test provider not found in stored archive")
	}
	data[at] ^= 1
	if err := os.WriteFile(u.archive, data, 0600); err != nil {
		t.Fatal(err)
	}
	if err := u.promote(); err == nil {
		t.Fatal("damaged archive accepted")
	}
	if len(*calls) != 0 {
		t.Fatalf("damaged archive touched service: %q", *calls)
	}
}

type releaseRoundTrip func(*http.Request) (*http.Response, error)

func (fn releaseRoundTrip) RoundTrip(request *http.Request) (*http.Response, error) {
	return fn(request)
}

func TestSignedEDRPrivateBoundedTransferBeforePromotion(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	archive, err := os.ReadFile(u.archive)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(u.archive); err != nil {
		t.Fatal(err)
	}
	u.url = func(string) string { return "https://release.invalid/native.zip" }
	u.client = func(bool) *http.Client {
		return &http.Client{Transport: releaseRoundTrip(func(request *http.Request) (*http.Response, error) {
			if request.URL.String() != "https://release.invalid/native.zip" {
				t.Errorf("unexpected release URL: %s", request.URL)
			}
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(archive)), ContentLength: int64(len(archive))}, nil
		})}
	}
	if err := u.download("ignored", false); err != nil {
		t.Fatal(err)
	}
	if err := release.CheckPrivateDirectory(u.downloadDir); err != nil || filepath.Dir(u.archive) != u.downloadDir {
		t.Fatalf("release was not staged privately: %s %v", u.archive, err)
	}
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err != nil {
		t.Fatal(err)
	}
	if len(*calls) == 0 || u.downloadDir != "" {
		t.Fatalf("download/promote state not consumed: %q dir=%q", *calls, u.downloadDir)
	}
}

func TestSignedEDRTransferRejectsOversizeBeforeDiskWrite(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	u.url = func(string) string { return "https://release.invalid/oversize.zip" }
	u.client = func(bool) *http.Client {
		return &http.Client{Transport: releaseRoundTrip(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(strings.NewReader("x")), ContentLength: maxEDRReleaseDownload + 1}, nil
		})}
	}
	if err := u.download("ignored", false); err == nil {
		t.Fatal("oversized transfer accepted")
	}
	if len(*calls) != 0 || u.downloadDir != "" {
		t.Fatal("oversized transfer touched service or retained staging")
	}
	dirs, err := filepath.Glob(filepath.Join(u.base, ".edr-download-*"))
	if err != nil || len(dirs) != 0 {
		t.Fatalf("oversized transfer retained private directory: %v %v", dirs, err)
	}
}

func TestSignedEDRRestartRecoversInterruptedPromotedPair(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(u.journalPath()); err != nil {
		t.Fatal("promoted pair has no recovery journal:", err)
	}
	restarted := *u
	restarted.backupDir, restarted.stageDir, restarted.downloadDir = "", "", ""
	restarted.backedUp, restarted.promotedNames, restarted.presentNames = nil, nil, nil
	restarted.promoted = false
	if err := restarted.recover(); err != nil {
		t.Fatal(err)
	}
	for name, want := range map[string]string{u.identity.ServiceName: "old service", u.identity.ProviderName: "old provider", u.identity.ServiceName + ".release.json": "old manifest"} {
		body, err := os.ReadFile(filepath.Join(u.base, name))
		if err != nil || string(body) != want {
			t.Fatalf("%s not restored after crash: %q %v", name, body, err)
		}
	}
	if _, err := os.Stat(u.journalPath()); !os.IsNotExist(err) {
		t.Fatalf("recovered journal retained: %v", err)
	}
	if strings.Join(*calls, ",") != "version,status,stop,status,start" {
		t.Fatalf("unexpected interrupted-update service sequence: %q", *calls)
	}
}

func TestSignedEDRRecoveryRejectsForgedJournalPaths(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	malformed := `{"schema":1,"backup_dir":"/tmp/unrelated","stage_dir":"/tmp/unrelated","present":[],"was_running":true}`
	if err := os.WriteFile(u.journalPath(), []byte(malformed), 0600); err != nil {
		t.Fatal(err)
	}
	if err := u.recover(); err == nil {
		t.Fatal("forged external rollback path accepted")
	}
	if len(*calls) != 0 {
		t.Fatalf("forged journal touched service: %q", *calls)
	}
}

func TestSignedEDRFreshInstallAndInterruptedRegistrationRecovery(t *testing.T) {
	for _, crash := range []bool{false, true} {
		t.Run(map[bool]string{false: "commit", true: "crash after registration"}[crash], func(t *testing.T) {
			u, calls := pairedReleaseUpdate(t, false)
			installed := false
			running := false
			u.status = func() (edrServiceState, error) {
				*calls = append(*calls, "status")
				return edrServiceState{installed: installed, running: running}, nil
			}
			u.install = func() error { *calls = append(*calls, "install"); installed = true; return nil }
			u.start = func() error { *calls = append(*calls, "start"); running = true; return nil }
			u.stop = func() error { *calls = append(*calls, "stop"); running = false; return nil }
			u.uninstall = func() error { *calls = append(*calls, "uninstall"); installed = false; return nil }
			if err := u.promote(); err != nil {
				t.Fatal(err)
			}
			if crash {
				if err := u.install(); err != nil {
					t.Fatal(err)
				}
				restarted := *u
				restarted.backupDir, restarted.stageDir = "", ""
				restarted.backedUp, restarted.promotedNames, restarted.presentNames = nil, nil, nil
				restarted.promoted = false
				if err := restarted.recover(); err != nil {
					t.Fatal(err)
				}
				if installed || !strings.Contains(strings.Join(*calls, ","), "uninstall") {
					t.Fatalf("interrupted new service registration retained: %q", *calls)
				}
				for _, name := range u.names {
					if _, err := os.Stat(filepath.Join(u.base, name)); !os.IsNotExist(err) {
						t.Fatalf("interrupted fresh file %q retained: %v", name, err)
					}
				}
			} else {
				if err := u.finish(); err != nil {
					t.Fatal(err)
				}
				if !installed || !running || u.verifyInstalled() != nil {
					t.Fatal("fresh authenticated service not installed")
				}
			}
		})
	}
}

func TestSignedEDRAdoptedFilesStartNewRegistrationAndCleanFailedStart(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, false)
	installed, running := false, false
	u.status = func() (edrServiceState, error) {
		*calls = append(*calls, "status")
		return edrServiceState{installed: installed, running: running}, nil
	}
	u.install = func() error { *calls = append(*calls, "install"); installed = true; return nil }
	u.start = func() error { *calls = append(*calls, "start"); running = true; return nil }
	u.stop = func() error { *calls = append(*calls, "stop"); running = false; return nil }
	u.uninstall = func() error { *calls = append(*calls, "uninstall"); installed = false; return nil }
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err != nil {
		t.Fatal(err)
	}
	// Simulate a signed pair left on disk when service registration disappeared.
	installed, running = false, false
	if err := u.finish(); err != nil {
		t.Fatal(err)
	}
	if !installed || !running || u.expectedServiceSHA == "" {
		t.Fatalf("adopted signed pair not running: installed=%v running=%v hash=%q calls=%q", installed, running, u.expectedServiceSHA, *calls)
	}
	installed, running = false, false
	u.ready = func(time.Time) error { return errors.New("local readiness failed") }
	if err := u.finish(); err == nil || !strings.Contains(err.Error(), "local readiness failed") {
		t.Fatalf("failed adopted registration accepted: %v", err)
	}
	if installed || running || u.verifyInstalled() != nil {
		t.Fatalf("failed adopted registration retained: installed=%v running=%v calls=%q", installed, running, *calls)
	}
}

func TestSignedEDRPartialPromotionRestoresWholePriorPair(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	u.version = func(path string) error {
		*calls = append(*calls, "version")
		return os.Remove(filepath.Join(filepath.Dir(path), u.identity.ProviderName))
	}
	if err := u.promote(); err == nil {
		t.Fatal("partial pair promotion succeeded")
	}
	for name, want := range map[string]string{u.identity.ServiceName: "old service", u.identity.ProviderName: "old provider", u.identity.ServiceName + ".release.json": "old manifest"} {
		body, err := os.ReadFile(filepath.Join(u.base, name))
		if err != nil || string(body) != want {
			t.Fatalf("partial promotion lost old %s: %q %v", name, body, err)
		}
	}
	if _, err := os.Stat(u.journalPath()); !os.IsNotExist(err) {
		t.Fatalf("successful rollback retained journal: %v", err)
	}
}

func TestSignedEDRServiceStatusFailurePreservesOldPair(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, true)
	u.status = func() (edrServiceState, error) {
		*calls = append(*calls, "status")
		return edrServiceState{}, errors.New("service manager unavailable")
	}
	if err := u.promote(); err == nil {
		t.Fatal("service manager error ignored")
	}
	if strings.Join(*calls, ",") != "version,status" {
		t.Fatalf("status failure interrupted service: %q", *calls)
	}
	for name, want := range map[string]string{u.identity.ServiceName: "old service", u.identity.ProviderName: "old provider"} {
		body, err := os.ReadFile(filepath.Join(u.base, name))
		if err != nil || string(body) != want {
			t.Fatalf("status failure replaced %s: %q %v", name, body, err)
		}
	}
}

func TestSignedEDRAdoptsExistingAuthenticatedPairWithoutServiceChange(t *testing.T) {
	u, calls := pairedReleaseUpdate(t, false)
	key, err := u.key()
	if err != nil {
		t.Fatal(err)
	}
	stage, _, err := release.Extract(u.archive, u.base, key, u.identity)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{u.identity.ServiceName, u.identity.ProviderName} {
		if err := os.Rename(filepath.Join(stage, name), filepath.Join(u.base, name)); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Rename(filepath.Join(stage, "release.json"), filepath.Join(u.base, u.identity.ServiceName+".release.json")); err != nil {
		t.Fatal(err)
	}
	os.RemoveAll(stage)
	u.status = func() (edrServiceState, error) {
		*calls = append(*calls, "status")
		return edrServiceState{installed: true, running: true}, nil
	}
	u.install = func() error { t.Fatal("adoption reinstalled service"); return nil }
	u.start = func() error { t.Fatal("adoption restarted service"); return nil }
	if err := u.finish(); err != nil {
		t.Fatal(err)
	}
	if strings.Join(*calls, ",") != "status" {
		t.Fatalf("authenticated adoption changed service: %q", *calls)
	}
}

func TestNativeEDRPublisherKeyFailsClosed(t *testing.T) {
	previous := EDRReleasePublicKeyHex
	defer func() { EDRReleasePublicKeyHex = previous }()
	for _, invalid := range []string{"", "not a key", strings.Repeat("0", 62), strings.Repeat("0", 64)} {
		EDRReleasePublicKeyHex = invalid
		if _, err := edrReleasePublicKey(); err == nil {
			t.Fatalf("invalid publisher key %q accepted", invalid)
		}
	}
	private := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{21}, ed25519.SeedSize))
	want := private.Public().(ed25519.PublicKey)
	EDRReleasePublicKeyHex = hex.EncodeToString(want)
	got, err := edrReleasePublicKey()
	if err != nil || !bytes.Equal(got, want) {
		t.Fatalf("valid build-time publisher key rejected: %v", err)
	}
}

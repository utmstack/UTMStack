//go:build windows && edr_acceptance

package dependency

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/utmstack/UTMStack/agent/edr/release"
	"golang.org/x/sys/windows"
)

func requireWindowsReleaseLab(t *testing.T) string {
	t.Helper()
	root := os.Getenv("EDR_ACCEPTANCE_RELEASE_BUNDLE")
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" || root == "" {
		t.Skip("requires the disposable native Windows signed-pair phase")
	}
	user, err := windows.GetCurrentProcessToken().GetTokenUser()
	if err != nil || !user.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
		t.Fatal("native signed-pair acceptance requires SYSTEM", err)
	}
	return root
}

// The worker uses the actual updater/SCM paths and dies after durable promotion.
// The public lab pin is the only key material passed to the new process.
func TestWindowsAcceptanceReleaseCrashWorker(t *testing.T) {
	if os.Getenv("EDR_RELEASE_CRASH_WORKER") != "1" {
		t.Skip("owned abrupt-exit worker only")
	}
	requireWindowsReleaseLab(t)
	EDRReleasePublicKeyHex = os.Getenv("EDR_RELEASE_LAB_PUBLIC_KEY")
	u := newReleaseEDRUpdate(os.Getenv("EDR_RELEASE_LAB_BASE"))
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	os.Exit(73)
}

func TestWindowsAcceptanceSignedRelease(t *testing.T) {
	root := requireWindowsReleaseLab(t)
	state, err := edrServiceStatus()
	if err != nil || state.installed {
		t.Fatal("refusing an existing SCM service", err)
	}
	_, key, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	previousPin := EDRReleasePublicKeyHex
	EDRReleasePublicKeyHex = hex.EncodeToString(key.Public().(ed25519.PublicKey))
	defer func() { EDRReleasePublicKeyHex = previousPin }()
	base := os.Getenv("EDR_ACCEPTANCE_RELEASE_BASE")
	if filepath.Dir(base) != filepath.Clean(root) || filepath.Base(base) != "owned-release-transaction" {
		t.Fatal("explicit owned release transaction directory required")
	}
	if err := os.Mkdir(base, 0700); err != nil {
		t.Fatal(err)
	}
	u := newReleaseEDRUpdate(base)
	// Bind the same real install command to this owned directory, rather than
	// the test executable's directory. All SCM, stop/start, readiness, journal
	// and installed-file verification paths retain the product implementation.
	u.install = func() error {
		cmd := exec.Command(filepath.Join(base, u.identity.ServiceName), "install")
		cmd.Dir = base
		output, err := cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("native install: %w: %s", err, output)
		}
		return nil
	}
	t.Cleanup(func() {
		_ = u.stop()
		_ = u.uninstall()
		state, err := u.status()
		if err != nil || state.installed {
			t.Errorf("owned SCM registration remains: %+v %v", state, err)
		}
	})
	source := t.TempDir()
	for name, input := range map[string]string{u.identity.ServiceName: filepath.Join(root, "utmstack_edr.exe"), u.identity.ProviderName: filepath.Join(root, "native", "utmstack_amsi.dll")} {
		data, err := os.ReadFile(input)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(source, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	alternate := t.TempDir()
	for _, name := range []string{u.identity.ServiceName, u.identity.ProviderName} {
		data, err := os.ReadFile(filepath.Join(os.Getenv("EDR_ACCEPTANCE_RELEASE_UPDATE"), name))
		if err != nil {
			t.Fatal(err)
		}
		original, err := os.ReadFile(filepath.Join(source, name))
		if err != nil || sha256.Sum256(data) == sha256.Sum256(original) {
			t.Fatal("native release controls need two distinct signed byte sets", name, err)
		}
		if err := os.WriteFile(filepath.Join(alternate, name), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	db := filepath.Join(base, "db")
	if err := os.Mkdir(db, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(db, "local.ndb"), []byte("UTM.Acceptance.Release:0:*:55544d535441434b5f414d53495f414343455054414e43455f32303236\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := map[string]any{"enabled": true, "signature_db_dir": db, "quarantine_dir": filepath.Join(base, "quarantine"), "sensors": map[string]bool{"file_watcher": false, "process_guard": false, "amsi": true, "behavioral": false}, "ransomware": map[string]bool{"enabled": false}, "blocklist": map[string]bool{"enabled": false}, "scheduled_scan": map[string]bool{"enabled": false}}
	data, _ := json.Marshal(cfg)
	if err := os.WriteFile(filepath.Join(base, "edr.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
	pack := func(update *releaseEDRUpdate, input string) {
		t.Helper()
		if err := os.Remove(update.archive); err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		if _, err := release.Pack(update.archive, input, key, update.identity); err != nil {
			t.Fatal(err)
		}
	}
	status := func(update *releaseEDRUpdate, since time.Time) edrRuntimeStatus {
		t.Helper()
		if err := update.verifyInstalled(); err != nil {
			t.Fatal(err)
		}
		if err := update.waitForReady(since); err != nil {
			t.Fatal(err)
		}
		for _, mode := range []string{"registered-owned", "registered"} {
			probe := exec.Command(filepath.Join(root, "native", "acceptance_probe.exe"), filepath.Join(base, update.identity.ProviderName), mode)
			output, err := probe.CombinedOutput()
			if err != nil {
				t.Fatalf("real registered COM/AMSI probe (%s): %v: %s", mode, err, output)
			}
			t.Logf("registered COM/AMSI (%s): %s", mode, output)
		}
		body, err := os.ReadFile(filepath.Join(base, "status.json"))
		if err != nil {
			t.Fatal(err)
		}
		var got edrRuntimeStatus
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatal(err)
		}
		if !got.SensorAMSI || !got.AMSILocalReady {
			t.Fatal("native AMSI readiness was not exercised")
		}
		return got
	}
	var receipts []map[string]any
	record := func(name string, got edrRuntimeStatus) {
		provider, err := os.ReadFile(filepath.Join(base, u.identity.ProviderName))
		if err != nil {
			t.Fatal(err)
		}
		receipts = append(receipts, map[string]any{"case": name, "instance_id": got.InstanceID, "process_id": got.ProcessID, "service_sha256": got.ServiceSHA256, "provider_sha256": fmt.Sprintf("%x", sha256.Sum256(provider)), "amsi_local_ready": got.AMSILocalReady, "registered_com_amsi_probe": true})
		t.Logf("case=%s instance=%s pid=%d sha256=%s AMSI=%v", name, got.InstanceID, got.ProcessID, got.ServiceSHA256, got.AMSILocalReady)
	}
	defer func() {
		if path := os.Getenv("EDR_ACCEPTANCE_RELEASE_RECEIPT"); path != "" {
			body, _ := json.MarshalIndent(map[string]any{"scope": "disposable lab Ed25519 pin and Authenticode; real SCM/AMSI; no production signing custody or power-loss qualification", "public_key": EDRReleasePublicKeyHex, "cases": receipts}, "", "  ")
			if err := os.WriteFile(path, body, 0600); err != nil {
				t.Error(err)
			}
		}
	}()
	pack(u, source)
	start := time.Now().UTC()
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	if err := u.finish(); err != nil {
		t.Fatal(err)
	}
	first := status(u, start)
	record("fresh-signed-install", first)

	// An archive-byte mutation must not stop the already healthy real service.
	bad := newReleaseEDRUpdate(base)
	pack(bad, alternate)
	archive, err := os.ReadFile(bad.archive)
	if err != nil {
		t.Fatal(err)
	}
	archive[len(archive)/2] ^= 1
	if err := os.WriteFile(bad.archive, archive, 0600); err != nil {
		t.Fatal(err)
	}
	if err := bad.promote(); err == nil {
		t.Fatal("altered release archive was promoted")
	}
	unchanged := status(u, start)
	if unchanged.InstanceID != first.InstanceID {
		t.Fatal("untrusted archive stopped the real service")
	}
	record("altered-archive-rejected-before-stop", unchanged)

	update := newReleaseEDRUpdate(base)
	pack(update, alternate)
	start = time.Now().UTC()
	if err := update.promote(); err != nil {
		t.Fatal(err)
	}
	if err := update.finish(); err != nil {
		t.Fatal(err)
	}
	second := status(update, start)
	if second.InstanceID == first.InstanceID || second.ServiceSHA256 == first.ServiceSHA256 {
		t.Fatal("signed update did not restart into a distinct signed byte set")
	}
	record("signed-pair-promote-and-restart", second)

	// Inject a failed post-start probe after the replacement really becomes
	// ready. This qualifies rollback of real running files/SCM/COM, not an OS
	// start failure, which remains a separate fault surface.
	rollback := newReleaseEDRUpdate(base)
	pack(rollback, source)
	rollback.ready = func(since time.Time) error {
		if err := rollback.waitForReady(since); err != nil {
			return err
		}
		return errors.New("owned injected downstream readiness failure")
	}
	start = time.Now().UTC()
	if err := rollback.promote(); err != nil {
		t.Fatal(err)
	}
	rollbackError := rollback.finish()
	if rollbackError == nil || !strings.Contains(rollbackError.Error(), "owned injected downstream readiness failure") {
		t.Fatal("intended readiness fault was not observed", rollbackError)
	}
	t.Logf("injected readiness/rollback result: %v", rollbackError)
	if _, err := os.Stat(rollback.journalPath()); !os.IsNotExist(err) {
		t.Fatalf("readiness rollback remains pending: %v; result: %v", err, rollbackError)
	}
	restored := status(u, start)
	if restored.ServiceSHA256 != second.ServiceSHA256 {
		t.Fatal("readiness rollback did not restore prior signed service bytes")
	}
	record("real-running-pair-readiness-rollback", restored)

	crash := newReleaseEDRUpdate(base)
	pack(crash, source)
	self, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(self, "-test.v", "-test.timeout=3m", "-test.run=^TestWindowsAcceptanceReleaseCrashWorker$")
	cmd.Env = append(os.Environ(), "EDR_RELEASE_CRASH_WORKER=1", "EDR_RELEASE_LAB_PUBLIC_KEY="+EDRReleasePublicKeyHex, "EDR_RELEASE_LAB_BASE="+base)
	output, err := cmd.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 73 {
		t.Fatalf("owned abrupt-exit worker: %v: %s", err, output)
	}
	if _, err := os.Stat(crash.journalPath()); err != nil {
		t.Fatal("worker left no durable recovery journal", err)
	}
	start = time.Now().UTC()
	recovered := newReleaseEDRUpdate(base)
	if err := recovered.recover(); err != nil {
		t.Fatal(err)
	}
	afterCrash := status(recovered, start)
	if afterCrash.ServiceSHA256 != second.ServiceSHA256 {
		t.Fatal("abrupt recovery did not restore prior signed service bytes")
	}
	record("abrupt-process-exit-and-native-recovery", afterCrash)
	if err := recovered.recover(); err != nil {
		t.Fatal("repeated recovery", err)
	}
	if _, err := os.Stat(crash.journalPath()); !os.IsNotExist(err) {
		t.Fatal("completed journal remains", err)
	}

	// Restart from installed local bytes with the transport and archive gone.
	_ = os.Remove(recovered.archive)
	if err := recovered.stop(); err != nil {
		t.Fatal(err)
	}
	start = time.Now().UTC()
	if err := recovered.start(); err != nil {
		t.Fatal(err)
	}
	final := status(recovered, start)
	body, err := os.ReadFile(filepath.Join(base, recovered.identity.ServiceName))
	if err != nil || fmt.Sprintf("%x", sha256.Sum256(body)) != final.ServiceSHA256 {
		t.Fatal("offline service byte identity", err)
	}
	record("offline-installed-pair-restart", final)
}

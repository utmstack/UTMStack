//go:build linux

package dependency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/kardianos/service"
	edrconfig "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/shared/fs"
	"golang.org/x/sys/unix"
)

const linuxInstalledRoot = "/opt/openedr-acceptance"
const linuxOutsideName = "/srv/openedr-unprotected-owned (deleted)"

func installedLinuxScope(t *testing.T) (string, string) {
	t.Helper()
	if os.Getenv("OPENEDR_NATIVE_SYSTEMD_ACCEPTANCE") != "1" {
		t.Skip("requires an explicitly owned isolated systemd guest")
	}
	host, _ := os.Hostname()
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" || host != "openedr-linux-owned" || os.Geteuid() != 0 || service.Platform() != "linux-systemd" || fs.GetExecutablePath() != linuxInstalledRoot {
		t.Fatal("native installation requires the owned guest and exact fixture install root")
	}
	input := os.Getenv("OPENEDR_NATIVE_RELEASE_INPUT")
	if input != "/mnt/openedr-input" {
		t.Fatal("native publisher fixtures require the owned read-only input mount")
	}
	var stat unix.Statfs_t
	if err := unix.Statfs(input, &stat); err != nil || stat.Flags&unix.ST_RDONLY == 0 {
		t.Fatalf("native fixture input is not read-only: %v", err)
	}
	return linuxInstalledRoot, input
}

// An exited systemd unit still exists and must remain recoverable. This owned
// service-state fixture deliberately exits without producing an EDR status.
func TestNativeLinuxFailedServiceState(t *testing.T) {
	installedLinuxScope(t)
	state, err := edrServiceStatus()
	if err != nil || state.installed {
		t.Fatalf("refusing a preexisting service: %+v %v", state, err)
	}
	path := "/etc/systemd/system/UTMStackEDR.service"
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = nativeSystemctl("stop", "UTMStackEDR.service")
		_ = nativeSystemctl("reset-failed", "UTMStackEDR.service")
		_ = os.Remove(path)
		_ = nativeSystemctl("daemon-reload")
	})
	_, writeErr := io.WriteString(f, "[Unit]\nDescription=Owned OpenEDR failed-state control\n[Service]\nType=simple\nExecStart=/bin/false\nRestart=no\n")
	if err := errors.Join(writeErr, f.Close()); err != nil {
		t.Fatal(err)
	}
	if err = nativeSystemctl("daemon-reload"); err != nil {
		t.Fatal(err)
	}
	if err = nativeSystemctl("start", "UTMStackEDR.service"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		out, showErr := exec.CommandContext(ctx, "systemctl", "show", "UTMStackEDR.service", "--property=ActiveState", "--value").CombinedOutput()
		cancel()
		if showErr == nil && strings.TrimSpace(string(out)) == "failed" {
			nativeInstalledRecord(t, "failed-unit-observed", string(out))
			break
		}
		if time.Now().After(deadline) {
			t.Fatalf("owned service did not fail: %q %v", out, showErr)
		}
		time.Sleep(50 * time.Millisecond)
	}
	state, err = edrServiceStatus()
	if err != nil || !state.installed || state.running {
		t.Fatalf("failed installed unit cannot be recovered: %+v %v", state, err)
	}
}

// This campaign invokes the agent's production EDR dependency callbacks and
// systemd implementation. Only the publisher is an owned local fixture; no
// service-control, readiness, version, signature or download callback is mocked.
func TestNativeLinuxInstalledSignedRelease(t *testing.T) {
	base, input := installedLinuxScope(t)
	key, err := edrReleasePublicKey()
	if err != nil {
		t.Fatal(err)
	}
	public, err := os.ReadFile(filepath.Join(input, "release-public.hex"))
	if err != nil || strings.TrimSpace(string(public)) != hex.EncodeToString(key) {
		t.Fatal("test agent build does not pin this fixture publisher")
	}
	u := newReleaseEDRUpdate(base)
	state, err := u.status()
	if err != nil || state.installed {
		t.Fatalf("refusing a preexisting service: %+v %v", state, err)
	}
	var selected atomic.Value
	selected.Store(filepath.Join(input, "good-release.zip"))
	for _, name := range []string{"good-release.zip", "wrong-key-release.zip", "broken-release.zip"} {
		info, err := os.Stat(filepath.Join(input, name))
		if err != nil || !info.Mode().IsRegular() {
			entries, _ := os.ReadDir(input)
			t.Fatalf("missing publisher fixture %s: %v; input entries: %v", name, err, entries)
		}
	}
	listener, err := net.Listen("tcp", "127.0.0.1:9001")
	if err != nil {
		t.Fatal(err)
	}
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Logf("publisher request %s %s selected=%s", r.Method, r.URL.String(), selected.Load().(string))
		if r.URL.Path != "/private/dependencies/agent/"+EDRFile("")+".release.zip" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, selected.Load().(string))
	})}
	serverDone := make(chan error, 1)
	go func() {
		serverDone <- server.ServeTLS(listener, filepath.Join(input, "server.crt"), filepath.Join(input, "server.key"))
	}()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		<-serverDone
	})
	protected := "/srv/openedr-protected"
	watch := "/srv/openedr-empty-watch"
	db := filepath.Join(base, "database")
	for _, path := range []string{protected, watch, db} {
		if err := os.MkdirAll(path, 0755); err != nil {
			t.Fatal(err)
		}
	}
	marker := []byte("OPENEDR_OWNED_INSTALLED_READINESS_MARKER_20261111")
	if err := os.WriteFile(filepath.Join(protected, "inert-marker"), marker, 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(protected, "harmless"), []byte("ordinary harmless installed service fixture"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(db, "owned.ndb"), []byte("Safe.Installed.Readiness:0:*:"+hex.EncodeToString(marker)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := edrconfig.Default()
	cfg.Enabled = true
	cfg.SignatureDBDir = db
	cfg.WatchVolumes = []string{watch}
	cfg.Ransomware.Enabled = false
	cfg.Blocklist.Enabled = false
	cfg.ScheduledScan.Enabled = false
	cfg.ScanConcurrency = 1
	on, off := true, false
	cfg.Sensors = edrconfig.Sensors{FileWatcher: &on, ProcessGuard: &off, AMSI: &off, Behavioral: &off}
	cfg.LinuxProtection.Fanotify = edrconfig.LinuxFanotify{Enabled: true, Mounts: []string{protected}, Roots: []string{protected}, Mode: "permission_open", FailClosed: true, DeadlineMS: 1500, Workers: 2, MaxPending: 16}
	if os.Getenv("OPENEDR_NATIVE_ROOT_MARK") == "1" {
		cfg.LinuxProtection.Fanotify.Mounts = []string{"/"}
	}
	if err = edrconfig.Save(cfg); err != nil {
		t.Fatal(err)
	}
	dropDir := "/etc/systemd/system/UTMStackEDR.service.d"
	if err = os.MkdirAll(dropDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(filepath.Join(dropDir, "90-owned-test-limits.conf"), []byte("[Service]\nMemoryMax=2G\nMemorySwapMax=0\nCPUQuota=200%\n"), 0644); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = nativeSystemctl("stop", "UTMStackEDR.service")
		fresh := newReleaseEDRUpdate(base)
		_ = fresh.recover()
		_ = nativeSystemctl("stop", "UTMStackEDR.service")
		_ = fresh.uninstall()
		_ = os.Remove(filepath.Join(dropDir, "90-owned-test-limits.conf"))
		_ = os.Remove(filepath.Join(dropDir, "91-owned-test-no-restart.conf"))
		_ = os.Remove(dropDir)
		_ = nativeSystemctl("daemon-reload")
	})
	dep := nativeEDRDependency(base)
	if err = dep.Download("127.0.0.1", false); err != nil {
		t.Fatal(err)
	}
	if err = dep.PostDownload(); err != nil {
		t.Fatal(err)
	}
	started := time.Now().UTC()
	if err = dep.Configure(); err != nil {
		t.Fatal(err)
	}
	fresh := newReleaseEDRUpdate(base)
	if err = fresh.verifyInstalled(); err != nil {
		t.Fatal(err)
	}
	if err = fresh.checkReady(started.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	before := nativeInstalledStatus(t, base)
	nativeInstalledRecord(t, "cold-install", before)
	for name, want := range map[string]string{"memory.max": "2147483648", "memory.swap.max": "0"} {
		data, err := os.ReadFile(filepath.Join("/sys/fs/cgroup/system.slice/UTMStackEDR.service", name))
		if err != nil || strings.TrimSpace(string(data)) != want {
			t.Fatalf("installed %s limit: %q %v", name, data, err)
		}
		nativeInstalledRecord(t, name, strings.TrimSpace(string(data)))
	}
	attr, err := os.ReadFile(fmt.Sprintf("/proc/%d/attr/current", int(before["process_id"].(float64))))
	if err != nil {
		t.Fatal(err)
	}
	nativeInstalledRecord(t, "apparmor-profile", strings.TrimSpace(string(attr)))
	serviceHash := nativeInstalledDigest(t, filepath.Join(base, EDRFile("")))
	manifestHash := nativeInstalledDigest(t, filepath.Join(base, EDRFile("")+".release.json"))
	for _, uid := range []uint32{0, 65534} {
		nativeInstalledRead(t, filepath.Join(protected, "harmless"), uid, "READ_ALLOWED")
		nativeInstalledRead(t, filepath.Join(protected, "inert-marker"), uid, "READ_DENIED")
	}
	if os.Getenv("OPENEDR_NATIVE_ROOT_MARK") == "1" {
		outsideAt := time.Now().UTC()
		if err = os.WriteFile(linuxOutsideName, []byte("inert excluded directory fixture"), 0644); err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = os.Remove(linuxOutsideName) })
		for _, uid := range []uint32{0, 65534} {
			nativeInstalledRead(t, linuxOutsideName, uid, "READ_ALLOWED")
		}
		deadline := time.Now().Add(40 * time.Second)
		for {
			after := nativeInstalledStatus(t, base)
			updated, _ := time.Parse(time.RFC3339Nano, fmt.Sprint(after["updated_at"]))
			if updated.After(outsideAt) {
				if err = fresh.checkReady(started.Add(-time.Second)); err != nil {
					t.Fatal(err)
				}
				nativeInstalledRecord(t, "outside-name-excluded", after)
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("outside-name control never reached a new status period")
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
	if err = nativeSystemctl("show", "UTMStackEDR.service", "--property=MemoryMax", "--property=MemorySwapMax", "--property=ControlGroup", "--property=MainPID"); err != nil {
		t.Fatal(err)
	}
	// A valid HTTPS mirror cannot authorize a release under another key. The
	// current running process and exact signed files must survive rejection.
	selected.Store(filepath.Join(input, "wrong-key-release.zip"))
	rejected := newReleaseEDRUpdate(base)
	if err = rejected.download("127.0.0.1", false); err != nil {
		t.Fatal(err)
	}
	if err = rejected.promote(); err == nil {
		t.Fatal("foreign publisher release accepted")
	}
	if after := nativeInstalledStatus(t, base); after["process_id"] != before["process_id"] {
		t.Fatal("signature rejection stopped existing protection")
	}
	if nativeInstalledDigest(t, filepath.Join(base, EDRFile(""))) != serviceHash || nativeInstalledDigest(t, filepath.Join(base, EDRFile("")+".release.json")) != manifestHash {
		t.Fatal("rejected publisher changed installed signed bytes")
	}
	nativeInstalledRecord(t, "wrong-key-refused", map[string]any{"service_pid": before["process_id"], "error": err.Error()})
	selected.Store(filepath.Join(input, "good-release.zip"))
	// Even with the engine and systemd service still healthy, losing the
	// permission owner must make this configured release unready.
	pid := int(before["process_id"].(float64))
	owners := nativeServiceChildren(t, pid)
	if len(owners) != 1 {
		t.Fatalf("expected sole independent owner: %v", owners)
	}
	if err = unix.Kill(owners[0], unix.SIGKILL); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(35 * time.Second)
	for {
		after := nativeInstalledStatus(t, base)
		sensors, _ := after["sensor_states"].(map[string]any)
		prevention, _ := sensors["file_prevention"].(map[string]any)
		if prevention["state"] == "degraded" {
			if after["engine_healthy"] != true {
				t.Fatal("fixture did not isolate protection readiness from engine health")
			}
			if err = fresh.checkReady(started.Add(-time.Second)); err == nil {
				t.Fatal("configured dead permission owner committed as healthy release")
			}
			nativeInstalledRecord(t, "owner-loss-readiness-refused", after)
			nativeInstalledRead(t, filepath.Join(protected, "inert-marker"), 0, "READ_ALLOWED")
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("owner loss never reached native status")
		}
		time.Sleep(100 * time.Millisecond)
	}
	restartAt := time.Now().UTC()
	if err = nativeSystemctl("restart", "UTMStackEDR.service"); err != nil {
		t.Fatal(err)
	}
	nativeWaitInstalled(t, base, restartAt)
	nativeInstalledRecord(t, "restart", nativeInstalledStatus(t, base))
	// A broken but authenticated program passes the version command and fails
	// real systemd startup. The updater must restore the previous signed bytes.
	if err = os.WriteFile(filepath.Join(dropDir, "91-owned-test-no-restart.conf"), []byte("[Service]\nRestart=no\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if err = nativeSystemctl("daemon-reload"); err != nil {
		t.Fatal(err)
	}
	selected.Store(filepath.Join(input, "broken-release.zip"))
	broken := newReleaseEDRUpdate(base)
	if err = broken.download("127.0.0.1", false); err != nil {
		t.Fatal(err)
	}
	if err = broken.promote(); err != nil {
		t.Fatal(err)
	}
	rollbackAt := time.Now().UTC()
	failure := broken.finish()
	if failure == nil {
		t.Fatal("broken signed service was committed")
	}
	nativeWaitInstalled(t, base, rollbackAt)
	if nativeInstalledDigest(t, filepath.Join(base, EDRFile(""))) != serviceHash || nativeInstalledDigest(t, filepath.Join(base, EDRFile("")+".release.json")) != manifestHash {
		t.Fatal("startup rollback did not restore exact prior signed files")
	}
	nativeInstalledRecord(t, "real-start-failure-rollback", map[string]any{"error": failure.Error(), "status": nativeInstalledStatus(t, base)})
	selected.Store(filepath.Join(input, "good-release.zip"))
	// This child exits after actual file promotion without committing, leaving
	// the production recovery journal and genuine stopped service behind.
	child := exec.Command(os.Args[0], "-test.run=^TestNativeLinuxReleaseCrashFixture$", "-test.timeout=30s")
	child.Env = append(os.Environ(), "OPENEDR_NATIVE_RELEASE_CRASH=1")
	out, err := child.CombinedOutput()
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 97 {
		t.Fatalf("owned promotion exit checkpoint: %v %s", err, out)
	}
	if _, err = os.Stat(filepath.Join(base, edrJournalFile)); err != nil {
		t.Fatal("abrupt promotion lost recovery evidence")
	}
	recoverAt := time.Now().UTC()
	recovering := nativeEDRDependency(base)
	if err = recovering.Recover(); err != nil {
		t.Fatal(err)
	}
	nativeWaitInstalled(t, base, recoverAt)
	nativeInstalledRecord(t, "process-crash-recovered", nativeInstalledStatus(t, base))
	if _, err = os.Stat(filepath.Join(base, edrJournalFile)); !os.IsNotExist(err) {
		t.Fatal("completed recovery retained transaction journal")
	}
	for _, name := range []string{"memory.current", "memory.peak", "memory.events", "cpu.stat"} {
		data, err := os.ReadFile(filepath.Join("/sys/fs/cgroup/system.slice/UTMStackEDR.service", name))
		if err != nil {
			t.Fatal(err)
		}
		nativeInstalledRecord(t, name, string(data))
	}
}

func TestNativeLinuxOwnedReadFixture(t *testing.T) {
	path := os.Getenv("OPENEDR_NATIVE_READ_PATH")
	if path == "" {
		return
	}
	host, _ := os.Hostname()
	allowed := path == "/srv/openedr-protected/inert-marker" || path == "/srv/openedr-protected/harmless" || (path == linuxOutsideName && os.Getenv("OPENEDR_NATIVE_ROOT_MARK") == "1")
	if os.Getenv("OPENEDR_ISOLATED_VM") != "1" || host != "openedr-linux-owned" || !allowed {
		t.Fatal("read actor accepts only owned inert guest fixtures")
	}
	f, err := os.Open(path)
	if os.IsPermission(err) {
		fmt.Println("READ_DENIED")
		return
	}
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err = io.Copy(io.Discard, io.LimitReader(f, 1<<20)); err != nil {
		t.Fatal(err)
	}
	fmt.Println("READ_ALLOWED")
}

func nativeInstalledRead(t *testing.T, path string, uid uint32, want string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestNativeLinuxOwnedReadFixture$")
	cmd.Env = append(os.Environ(), "OPENEDR_NATIVE_READ_PATH="+path)
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: uid, Gid: uid}}
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), want) {
		t.Fatalf("installed service uid=%d read=%s want=%s: %s %v", uid, path, want, out, err)
	}
	nativeInstalledRecord(t, "read-effect", map[string]any{"path": path, "uid": uid, "effect": want})
}

func TestNativeLinuxReleaseCrashFixture(t *testing.T) {
	if os.Getenv("OPENEDR_NATIVE_RELEASE_CRASH") != "1" {
		return
	}
	base, _ := installedLinuxScope(t)
	u := newReleaseEDRUpdate(base)
	if err := u.download("127.0.0.1", false); err != nil {
		t.Fatal(err)
	}
	if err := u.promote(); err != nil {
		t.Fatal(err)
	}
	os.Exit(97)
}

func nativeWaitInstalled(t *testing.T, base string, start time.Time) {
	t.Helper()
	u := newReleaseEDRUpdate(base)
	if err := u.verifyInstalled(); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(40 * time.Second)
	var last error
	for time.Now().Before(deadline) {
		last = u.checkReady(start)
		if last == nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
	t.Fatalf("installed service did not recover ready state: %v", last)
}

func nativeInstalledStatus(t *testing.T, base string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(base, "status.json"))
	if err != nil {
		t.Fatal(err)
	}
	var status map[string]any
	if err = json.Unmarshal(data, &status); err != nil {
		t.Fatal(err)
	}
	return status
}

func nativeServiceChildren(t *testing.T, pid int) []int {
	t.Helper()
	tasks, err := os.ReadDir(fmt.Sprintf("/proc/%d/task", pid))
	if err != nil {
		t.Fatal(err)
	}
	set := map[int]bool{}
	for _, task := range tasks {
		data, _ := os.ReadFile(fmt.Sprintf("/proc/%d/task/%s/children", pid, task.Name()))
		for _, s := range strings.Fields(string(data)) {
			id, e := strconv.Atoi(s)
			if e != nil {
				t.Fatal(e)
			}
			set[id] = true
		}
	}
	var children []int
	for id := range set {
		children = append(children, id)
	}
	return children
}

func nativeSystemctl(args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "systemctl", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("systemctl %v: %v %s", args, err, out)
	}
	return nil
}

func nativeInstalledRecord(t *testing.T, phase string, value any) {
	t.Helper()
	data, err := json.Marshal(map[string]any{"phase": phase, "value": value, "at": time.Now().UTC()})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("NATIVE_LINUX_RESULT %s", data)
}

func nativeInstalledDigest(t *testing.T, path string) string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(h.Sum(nil))
}

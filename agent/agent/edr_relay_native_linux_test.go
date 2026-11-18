//go:build linux

package agent

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/threatwinds/go-sdk/plugins"
	"github.com/utmstack/UTMStack/agent/edr/cache"
	edrcfg "github.com/utmstack/UTMStack/agent/edr/config"
	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/utils"
	"github.com/utmstack/UTMStack/shared/fs"
)

const nativeSpoolBase = "/opt/openedr-spool"

func spoolStatus() map[string]any {
	raw, _ := os.ReadFile(edrcfg.StatusFile)
	var s map[string]any
	_ = json.Unmarshal(raw, &s)
	return s
}
func waitSpoolStatus(t *testing.T, label string, pid int, predicate func(map[string]any) bool) map[string]any {
	t.Helper()
	deadline := time.Now().Add(38 * time.Second)
	for {
		s := spoolStatus()
		if s["process_id"] == float64(pid) && predicate(s) {
			raw, _ := json.Marshal(map[string]any{"phase": label, "status": s})
			t.Logf("NATIVE_SPOOL_STATUS %s", raw)
			return s
		}
		if time.Now().After(deadline) {
			t.Fatalf("%s status condition failed: %+v", label, s)
		}
		time.Sleep(25 * time.Millisecond)
	}
}
func waitSpoolCondition(t *testing.T, label string, timeout time.Duration, check func() bool) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for !check() {
		if time.Now().After(deadline) {
			t.Fatal(label)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
func TestNativeEDRRelayServiceBackpressureAndRestart(t *testing.T) {
	if os.Getenv("OPENEDR_NATIVE_SPOOL") != "1" {
		t.Skip("requires an owned offline Linux service fixture")
	}
	marker, err := os.ReadFile("/run/openedr-disposable-lab")
	if err != nil || string(marker) != "OPENEDR_DISPOSABLE_LAB_ONLY\n" || os.Geteuid() != 0 || fs.GetExecutablePath() != nativeSpoolBase {
		t.Fatal("not an owned Linux fixture")
	}
	utils.InitLogger(filepath.Join(nativeSpoolBase, "relay.log"))
	cfg := edrcfg.Default()
	cfg.Enabled = true
	cfg.ScanConcurrency = 1
	cfg.Ransomware.Enabled = false
	cfg.Blocklist.Enabled = false
	cfg.ScheduledScan.Enabled = false
	on, off := true, false
	cfg.Sensors = edrcfg.Sensors{FileWatcher: &on, ProcessGuard: &off, AMSI: &off, Behavioral: &off}
	cfg.SignatureDBDir = filepath.Join(nativeSpoolBase, "owned-db")
	cfg.WatchVolumes = []string{"/srv/openedr-spool-inbox"}
	for _, p := range []string{cfg.SignatureDBDir, cfg.WatchVolumes[0]} {
		if err = os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}
	inert := []byte("OPENEDR-INERT-SPOOL-CONTAINMENT-20261112")
	if err = os.WriteFile(filepath.Join(cfg.SignatureDBDir, "owned.ndb"), []byte("Lab.Spool.Containment:0:*:"+hex.EncodeToString(inert)+"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err = edrcfg.Save(cfg); err != nil {
		t.Fatal(err)
	}
	initialConfig, err := os.ReadFile(edrcfg.ConfigFile)
	if err != nil {
		t.Fatal(err)
	}
	configHash := sha256.Sum256(initialConfig)
	previousQueue := LogQueue
	LogQueue = make(chan *plugins.Log, 64)
	t.Cleanup(func() { LogQueue = previousQueue })
	for i := 0; i < cap(LogQueue); i++ {
		LogQueue <- &plugins.Log{DataType: "owned_fixture_hold", Raw: "owned fixture holds queue"}
	}
	relayCtx, cancelRelay := context.WithCancel(context.Background())
	relayDone := make(chan struct{})
	go func() { defer close(relayDone); EDRRelay(relayCtx) }()
	t.Cleanup(func() {
		cancelRelay()
		select {
		case <-relayDone:
		case <-time.After(2 * time.Second):
			t.Error("relay did not stop")
		}
	})
	var child *exec.Cmd
	var childDone chan error
	var childLog *os.File
	start := func() int {
		childLog, err = os.Create(filepath.Join(nativeSpoolBase, fmt.Sprintf("service-%d.log", time.Now().UnixNano())))
		if err != nil {
			t.Fatal(err)
		}
		child = exec.Command(filepath.Join(nativeSpoolBase, "utmstack_edr_linux_arm64"))
		child.Stdout = childLog
		child.Stderr = childLog
		if err = child.Start(); err != nil {
			t.Fatal(err)
		}
		childDone = make(chan error, 1)
		go func(c *exec.Cmd, ch chan error) { ch <- c.Wait() }(child, childDone)
		return child.Process.Pid
	}
	stop := func() {
		if child == nil {
			return
		}
		_ = child.Process.Signal(syscall.SIGTERM)
		select {
		case e := <-childDone:
			if e != nil {
				t.Errorf("service exit: %v", e)
			}
		case <-time.After(7 * time.Second):
			_ = child.Process.Kill()
			<-childDone
			t.Error("service required forced termination")
		}
		_ = childLog.Close()
		child = nil
	}
	t.Cleanup(stop)
	pid := start()
	ready := func(s map[string]any) bool {
		states, _ := s["sensor_states"].(map[string]any)
		w, _ := states["file_watcher"].(map[string]any)
		return s["running"] == true && s["engine_healthy"] == true && w["state"] == "active"
	}
	initial := waitSpoolStatus(t, "started", pid, ready)
	firstInstance := initial["instance_id"]
	sp, err := event.OpenSpool(edrcfg.SpoolFile, 8<<20)
	if err != nil {
		t.Fatal(err)
	}
	defer sp.Close()
	// The actual relay opens the v2 segment while the real bounded LogQueue is
	// deliberately full. No remote transport or fabricated delivery ack is used.
	waitSpoolCondition(t, "relay did not negotiate generation format", 3*time.Second, func() bool {
		f, e := os.Open(edrcfg.SpoolFile)
		if e != nil {
			return false
		}
		defer f.Close()
		var b [20]byte
		n, _ := f.Read(b[:])
		return strings.HasPrefix(string(b[:n]), `{"_openedr_spool":`)
	})
	accepted := map[int]bool{}
	full := 0
	for i := 0; i < 300; i++ {
		raw, e := json.Marshal(map[string]any{"product": event.Product, "engine": event.Product, "source": "behavioral", "action": "health", "owned_burst_ordinal": i, "padding": strings.Repeat("x", 64<<10)})
		if e != nil {
			t.Fatal(e)
		}
		e = sp.Append(string(raw))
		if e == nil {
			accepted[i] = true
		} else if errors.Is(e, event.ErrSpoolFull) {
			full++
		} else {
			t.Fatal(e)
		}
	}
	if full == 0 || len(accepted) < 200 {
		t.Fatalf("capacity control not exercised: accepted=%d full=%d", len(accepted), full)
	}
	// Large records can leave enough tail space for a short detection. Fill
	// that tail too, rather than weakening the actual service refusal check.
	for i := 300; i < 2300; i++ {
		raw, e := json.Marshal(map[string]any{"product": event.Product, "engine": event.Product, "source": "behavioral", "action": "health", "owned_burst_ordinal": i})
		if e != nil {
			t.Fatal(e)
		}
		e = sp.Append(string(raw))
		if errors.Is(e, event.ErrSpoolFull) {
			full++
			break
		}
		if e != nil {
			t.Fatal(e)
		}
		accepted[i] = true
		if i == 2299 {
			t.Fatal("small-record tail did not reach capacity")
		}
	}
	sizes := map[string]int64{}
	for _, name := range []string{edrcfg.SpoolFile, edrcfg.SpoolFile + ".1"} {
		info, e := os.Stat(name)
		if e != nil {
			t.Fatal(e)
		}
		sizes[filepath.Base(name)] = info.Size()
		if info.Size() > (8<<20)+128 {
			t.Fatal("segment exceeded bounded size")
		}
	}
	boundRaw, _ := json.Marshal(map[string]any{"accepted": len(accepted), "full_rejections": full, "segment_sizes": sizes})
	t.Logf("NATIVE_SPOOL_PRESSURE %s", boundRaw)
	contained := func(name string) {
		target := filepath.Join(cfg.WatchVolumes[0], name)
		if err := os.WriteFile(target, inert, 0600); err != nil {
			t.Fatal(err)
		}
		waitSpoolCondition(t, "real watcher did not contain "+name, 4*time.Second, func() bool { _, e := os.Stat(target); return os.IsNotExist(e) })
	}
	contained("during-full.txt")
	fullState := waitSpoolStatus(t, "full", pid, func(s map[string]any) bool {
		v, _ := s["event_spool"].(map[string]any)
		return v["degraded"] == true && v["full_rejections"].(float64) >= 1
	})
	var mu sync.Mutex
	received := map[int]bool{}
	detected := map[string]bool{}
	var receiveErr error
	receiveCtx, cancelReceive := context.WithCancel(context.Background())
	receiveDone := make(chan struct{})
	go func() {
		defer close(receiveDone)
		for {
			select {
			case <-receiveCtx.Done():
				return
			case log := <-LogQueue:
				if log.DataType == "owned_fixture_hold" {
					continue
				}
				var row map[string]any
				e := json.Unmarshal([]byte(log.Raw), &row)
				mu.Lock()
				if e != nil || row["product"] != event.Product {
					receiveErr = fmt.Errorf("non-event relayed: %v", e)
				}
				if n, ok := row["owned_burst_ordinal"].(float64); ok {
					if received[int(n)] {
						receiveErr = fmt.Errorf("duplicate ordinal %d", int(n))
					}
					received[int(n)] = true
				}
				if row["signature"] == "Lab.Spool.Containment" {
					detected[filepath.Base(fmt.Sprint(row["object_path"]))] = true
				}
				mu.Unlock()
			}
		}
	}()
	t.Cleanup(func() { cancelReceive(); <-receiveDone })
	recoveryStart := time.Now()
	waitSpoolCondition(t, "accepted burst records did not recover", 35*time.Second, func() bool { mu.Lock(); defer mu.Unlock(); return receiveErr != nil || len(received) == len(accepted) })
	mu.Lock()
	if receiveErr != nil {
		t.Error(receiveErr)
	}
	for n := range received {
		if !accepted[n] {
			t.Errorf("unaccepted ordinal delivered: %d", n)
		}
	}
	mu.Unlock()
	recoverySeconds := time.Since(recoveryStart).Seconds()
	contained("after-recovery.txt")
	waitSpoolCondition(t, "post-recovery detection did not reach actual LogQueue", 5*time.Second, func() bool { mu.Lock(); defer mu.Unlock(); return detected["after-recovery.txt"] })
	recovered := waitSpoolStatus(t, "recovered", pid, func(s map[string]any) bool {
		v, _ := s["event_spool"].(map[string]any)
		return v["degraded"] == true && v["accepted"].(float64) > fullState["event_spool"].(map[string]any)["accepted"].(float64) && v["full_rejections"].(float64) >= 1
	})
	stop()
	stopped := spoolStatus()
	st, _ := stopped["event_spool"].(map[string]any)
	if stopped["running"] != false || st["closed"] != true || st["full_rejections"].(float64) < 1 {
		t.Fatalf("final spool status lost: %+v", stopped)
	}
	stoppedRaw, _ := json.Marshal(stopped)
	t.Logf("NATIVE_SPOOL_STOPPED %s", stoppedRaw)
	pid = start()
	restarted := waitSpoolStatus(t, "restarted", pid, func(s map[string]any) bool {
		v, _ := s["event_spool"].(map[string]any)
		return ready(s) && s["instance_id"] != firstInstance && v["append_failures"] == float64(0)
	})
	contained("after-restart.txt")
	waitSpoolCondition(t, "restart detection did not reach LogQueue", 5*time.Second, func() bool { mu.Lock(); defer mu.Unlock(); return detected["after-restart.txt"] })
	stop()
	c, err := cache.Open(edrcfg.DBFile)
	if err != nil {
		t.Fatal(err)
	}
	rows, err := c.ListQuarantine()
	_ = c.Close()
	if err != nil || len(rows) != 3 {
		t.Fatalf("independent quarantines=%d err=%v", len(rows), err)
	}
	sum := sha256.Sum256(inert)
	for _, row := range rows {
		raw, e := os.ReadFile(filepath.Join(cfg.QuarantineDir, row.QuarantineID+".quarantined"))
		if e != nil || sha256.Sum256(raw) != sum || row.SHA256 != hex.EncodeToString(sum[:]) {
			t.Fatalf("quarantine snapshot mismatch: %+v %v", row, e)
		}
	}
	finalConfig, e := os.ReadFile(edrcfg.ConfigFile)
	if e != nil || sha256.Sum256(finalConfig) != configHash {
		t.Fatal("configuration changed during workload")
	}
	raw, _ := json.Marshal(map[string]any{"synthetic_accepted": len(accepted), "synthetic_full_rejections": full, "actual_quarantines": len(rows), "accepted_recovery_seconds": recoverySeconds, "actual_edr_relay": true, "full_status": fullState["event_spool"], "recovered_status": recovered["event_spool"], "restart_status": restarted["event_spool"], "configuration_sha256": fmt.Sprintf("%x", configHash), "remote_delivery_claimed": false})
	t.Logf("NATIVE_SPOOL_RESULT %s", raw)
}

//go:build windows && edr_acceptance

package ransomware

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/0xrawsec/golang-etw/etw"
	"github.com/utmstack/UTMStack/agent/edr/responder"
	"golang.org/x/sys/windows"
)

func requireETWAcceptanceSystem(t *testing.T) {
	t.Helper()
	if os.Getenv("EDR_WINDOWS_ACCEPTANCE") != "isolated-vm" {
		t.Skip("isolated Windows VM only")
	}
	var token windows.Token
	if err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token); err != nil {
		t.Fatal(err)
	}
	defer token.Close()
	user, err := token.GetTokenUser()
	if err != nil || !user.User.Sid.IsWellKnown(windows.WinLocalSystemSid) {
		t.Fatal("SYSTEM identity required for protected ETW ownership")
	}
}

func TestWindowsAcceptanceETWOrphanOwnerChild(t *testing.T) {
	ready := os.Getenv("EDR_WINDOWS_ETW_OWNER_READY")
	if ready == "" {
		return
	}
	requireETWAcceptanceSystem(t)
	s, err := startFileTrace(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	data, err := json.Marshal(s.owner)
	if err != nil {
		t.Fatal(err)
	}
	if err = os.WriteFile(ready+".incoming", data, 0600); err != nil {
		t.Fatal(err)
	}
	if err = os.Rename(ready+".incoming", ready); err != nil {
		t.Fatal(err)
	}
	// Parent kills this exact owned process after observing durable ownership.
	// This fallback is bounded if the parent unexpectedly exits first.
	time.Sleep(90 * time.Second)
}

func acceptanceCrashedTraceOwner(t *testing.T) traceOwnerRecord {
	t.Helper()
	requireETWAcceptanceSystem(t)
	ready := filepath.Join(t.TempDir(), "owned-trace-ready.json")
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	child := exec.Command(exe, "-test.run=^TestWindowsAcceptanceETWOrphanOwnerChild$", "-test.timeout=2m")
	child.Env = append(os.Environ(), "EDR_WINDOWS_ETW_OWNER_READY="+ready)
	childLog, err := os.Create(filepath.Join(filepath.Dir(ready), "child.log"))
	if err != nil {
		t.Fatal(err)
	}
	child.Stdout, child.Stderr = childLog, childLog
	t.Cleanup(func() {
		childLog.Close()
		if b, e := os.ReadFile(childLog.Name()); e == nil {
			if len(b) > 65536 {
				b = b[:65536]
			}
			t.Logf("owned helper output: %s", b)
		}
	})
	if err = child.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = child.Process.Kill()
		_ = child.Wait()
		ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
		defer cancel()
		store, e := openTraceOwnerStore(ctx)
		if e == nil {
			e = store.cleanupOrphans()
			store.Close()
		}
		if e != nil {
			t.Errorf("owned orphan fixture cleanup: %v", e)
		}
	})
	identity, err := (responder.OSKiller{}).ProcessIdentity(child.Process.Pid)
	if err != nil {
		t.Fatal(err)
	}
	var data []byte
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		data, err = os.ReadFile(ready)
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if err != nil {
		t.Fatal("owned trace child did not become ready", err)
	}
	var record traceOwnerRecord
	if err = json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	if err = record.validate(); err != nil {
		t.Fatal(err)
	}
	if record.PID != uint32(child.Process.Pid) || record.Creation != identity {
		t.Fatal("trace record does not identify the exact owned child")
	}
	native, err := queryOwnedTrace(record.Name, 0)
	if err != nil || !native.Exists || native.GUID != record.GUID || native.Name != record.Name {
		t.Fatalf("owned native trace missing: %+v %v", native, err)
	}
	if err = child.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	if err = child.Wait(); err == nil {
		t.Fatal("owned child did not terminate abruptly")
	}
	if traceProcessState(record) != traceProcessGone && traceProcessState(record) != traceProcessReused {
		t.Fatal("kernel did not prove the old owner gone")
	}
	return record
}

func TestWindowsAcceptanceETWCrashRestartReclaimsOnlyRecordedOwner(t *testing.T) {
	record := acceptanceCrashedTraceOwner(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, err := startFileTrace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	old, err := queryOwnedTrace(record.Name, 0)
	if err != nil || old.Exists {
		t.Fatalf("owned orphan survived verified cleanup: %+v %v", old, err)
	}
	f, err := s.store.openRecord(record.filename(), false)
	if err == nil {
		f.Close()
		t.Fatal("reclaimed orphan record remains")
	}
	if !errors.Is(err, windows.ERROR_FILE_NOT_FOUND) {
		t.Fatal(err)
	}
	t.Log("actual abrupt owned-process death, exact GUID cleanup and replacement session passed")
}

func TestWindowsAcceptanceETWForeignGUIDIsNeverStopped(t *testing.T) {
	record := acceptanceCrashedTraceOwner(t)
	native, err := queryOwnedTrace(record.Name, 0)
	if err != nil {
		t.Fatal(err)
	}
	if err = stopVerifiedOrphan(record, native); err != nil {
		t.Fatal(err)
	}
	guid, err := windows.GenerateGUID()
	if err != nil {
		t.Fatal(err)
	}
	foreignGUID := strings.ToUpper(guid.String())
	properties := fileTraceProperties(record.Name)
	properties.Wnode.Guid = *etw.MustParseGUIDFromString(foreignGUID)
	name, _ := windows.UTF16PtrFromString(record.Name)
	var handle syscall.Handle
	if err = etw.StartTrace(&handle, name, properties); err != nil {
		t.Fatal(err)
	}
	// This second session is an explicitly owned fixture, with no provider. Only
	// this local handle is used to dispose of it after refusal is verified.
	t.Cleanup(func() {
		if err := etw.ControlTrace(handle, nil, fileTraceProperties(record.Name), etw.EVENT_TRACE_CONTROL_STOP); err != nil {
			t.Errorf("foreign fixture teardown: %v", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if s, e := startFileTrace(ctx); e == nil {
		s.Stop()
		t.Fatal("mismatched native GUID was accepted")
	}
	after, err := queryOwnedTrace(record.Name, 0)
	if err != nil || !after.Exists || after.GUID != foreignGUID {
		t.Fatalf("foreign fixture was stopped or changed: %+v %v", after, err)
	}
	t.Log("same-name/different-GUID native session preserved; ownership remains degraded")
}

func TestWindowsAcceptanceETWCallbackHealthAndCloseJoin(t *testing.T) {
	requireETWAcceptanceSystem(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	s, err := startFileTrace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Stop()
	observed := make(chan struct{}, 1)
	expected := errors.New("controlled native callback failure")
	c, err := openFileConsumer(ctx, s.name, func(consumerCtx context.Context, record *etw.EventRecord) error {
		if consumerCtx.Err() != nil {
			return consumerCtx.Err()
		}
		if record.EventHeader.ProviderId.Equals(etw.MustParseGUIDFromString(kernelFileProviderGUID)) {
			select {
			case observed <- struct{}{}:
			default:
			}
			return expected
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer c.Stop()
	fixture := filepath.Join(t.TempDir(), "inert-native-health.txt")
	for {
		if err = os.WriteFile(fixture, []byte("ordinary controlled ETW I/O"), 0600); err != nil {
			t.Fatal(err)
		}
		select {
		case <-observed:
			goto delivered
		case <-ctx.Done():
			t.Fatal("native ETW callback was not delivered")
		case <-time.After(20 * time.Millisecond):
		}
	}
delivered:
	deadline := time.Now().Add(time.Second)
	for c.Err() == nil && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond)
	}
	if c.Err() != expected {
		t.Fatal("callback failure was not retained", c.Err())
	}
	cancel()
	var group sync.WaitGroup
	for i := 0; i < 8; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			if e := c.Stop(); e != nil {
				t.Error(e)
			}
			if c.Err() != expected {
				t.Error("close lost callback health")
			}
		}()
	}
	group.Wait()
	if err = s.Stop(); err != nil {
		t.Fatal(err)
	}
	native, err := queryOwnedTrace(s.name, 0)
	if err != nil || native.Exists {
		t.Fatalf("native trace survived normal stop: %+v %v", native, err)
	}
	t.Log("actual native callback failure, concurrent health/close and joined shutdown passed")
}

func TestWindowsAcceptanceETWConcurrentOwnerCleanupAndStop(t *testing.T) {
	requireETWAcceptanceSystem(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	session, err := startFileTrace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Stop()
	other, err := openTraceOwnerStore(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer other.Close()
	stopped := make(chan error, 1)
	go func() { stopped <- session.Stop() }()
	// Startup owns the interprocess lock while opening protected DELETE record
	// handles; concurrent normal Stop must wait rather than memoize a share error.
	if err = other.cleanupOrphans(); err != nil {
		t.Fatal(err)
	}
	other.Close()
	select {
	case err = <-stopped:
		if err != nil {
			t.Fatal(err)
		}
	case <-ctx.Done():
		t.Fatal("concurrent ownership cleanup and stop did not join")
	}
	t.Log("protected startup cleanup and normal stop serialized without record-sharing loss")
}

func TestWindowsAcceptanceETWHardLinkDoesNotBecomeDestructiveRename(t *testing.T) {
	requireETWAcceptanceSystem(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	folder := t.TempDir()
	original, alias := filepath.Join(folder, "inert-decoy.txt"), filepath.Join(folder, "inert-alias.txt")
	if err := os.WriteFile(original, []byte("ordinary hard-link control"), 0600); err != nil {
		t.Fatal(err)
	}
	session, err := startFileTrace(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer session.Stop()
	decoder := newKernelFileDecoder(func(path string) bool { return strings.EqualFold(path, original) })
	opened := make(chan struct{}, 1)
	linked := make(chan kernelFileRecord, 1)
	destructive := make(chan FileEvent, 8)
	provider := etw.MustParseGUIDFromString(kernelFileProviderGUID)
	var diagnosticMu sync.Mutex
	ids := map[uint16]int{}
	pids := map[uint32]int{}
	paths := []string{}
	t.Cleanup(func() {
		diagnosticMu.Lock()
		defer diagnosticMu.Unlock()
		t.Logf("bounded hard-link native diagnostics: ids=%v pids=%v own_pid=%d owned_paths=%q expected=%q", ids, pids, os.Getpid(), paths, original)
	})
	consumer, err := openFileConsumer(ctx, session.name, func(callbackCtx context.Context, record *etw.EventRecord) error {
		if callbackCtx.Err() != nil || !record.EventHeader.ProviderId.Equals(provider) {
			return nil
		}
		id := record.EventHeader.EventDescriptor.Id
		diagnosticMu.Lock()
		if len(ids) < 64 || ids[id] > 0 {
			ids[id]++
		}
		if len(pids) < 32 || pids[record.EventHeader.ProcessId] > 0 {
			pids[record.EventHeader.ProcessId]++
		}
		diagnosticMu.Unlock()
		if record.EventHeader.ProcessId != uint32(os.Getpid()) {
			return nil
		}
		if !kernelFileEventWanted(id) && id != kfSetLink && id != 28 {
			return nil
		}
		native, e := nativeKernelFileRecord(record)
		if e != nil {
			return e
		}
		if native.Path != "" {
			diagnosticMu.Lock()
			if len(paths) < 16 {
				path := native.Path
				if len(path) > 512 {
					path = path[:512]
				}
				paths = append(paths, path)
			}
			diagnosticMu.Unlock()
		}
		if id == kfCreate && strings.EqualFold(native.Path, original) {
			select {
			case opened <- struct{}{}:
			default:
			}
		}
		if id == kfSetLink || id == 28 {
			// Exercise the former SetLink-as-rename bug on an actually tracked
			// canary FileObject, rather than accepting an unrelated link record.
			if strings.EqualFold(decoder.opens[native.FileObject], original) {
				select {
				case linked <- native:
				default:
				}
			}
		}
		if event, ok := decoder.decode(native); ok {
			select {
			case destructive <- event:
			default:
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer consumer.Stop()
	// ProcessTrace starts asynchronously. Repeat only harmless read-opens until
	// native delivery proves this exact path is visible; do not infer readiness
	// from a successful OpenTrace handle or an otherwise empty health error.
	var handle *os.File
	defer func() {
		if handle != nil {
			handle.Close()
		}
	}()
readReady:
	for {
		if handle != nil {
			handle.Close()
		}
		handle, err = os.Open(original)
		if err != nil {
			t.Fatal(err)
		}
		select {
		case <-opened:
			break readReady
		case <-ctx.Done():
			t.Fatal("owned read-open was not delivered", consumer.Err())
		case <-time.After(25 * time.Millisecond):
		}
	}
	if err = os.Link(original, alias); err != nil {
		t.Fatal(err)
	}
	originalInfo, err := os.Stat(original)
	if err != nil {
		t.Fatal(err)
	}
	aliasInfo, err := os.Stat(alias)
	if err != nil || !os.SameFile(originalInfo, aliasInfo) {
		t.Fatal("owned hard-link fixture does not identify the original file", err)
	}
	select {
	case event := <-destructive:
		t.Fatal("hard-link control became a destructive event", event)
	case record := <-linked:
		t.Logf("actual owned hard-link native event id=%d file_object=%d correlated_decoy=%q", record.EventID, record.FileObject, original)
	case <-ctx.Done():
		t.Fatal("owned hard-link event was not delivered", consumer.Err())
	}
	// The native flush interval is one second; retain a bounded observation window
	// after the actual hard-link event, without installing a response/kill sink.
	timer := time.NewTimer(1500 * time.Millisecond)
	defer timer.Stop()
	select {
	case event := <-destructive:
		t.Fatal("hard-link control became a destructive event", event)
	case <-timer.C:
	}
	if err = consumer.Err(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{original, alias} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != "ordinary hard-link control" {
			t.Fatal("hard-link control bytes changed", err)
		}
	}
	t.Log("actual harmless hard-link preserved existing canary content and emitted no destructive decoder operation")
}

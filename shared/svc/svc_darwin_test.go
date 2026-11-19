//go:build darwin

package svc

import (
	"fmt"
	"reflect"
	"testing"
)

// darwinStub is a fake command runner + plist-installed seam for the darwin
// service layer. It records every command and lets a test drive the exit
// code/output of `launchctl list` (the "loaded" signal) and force failures on
// a specific verb (load/unload).
type darwinStub struct {
	calls      [][]string
	listLoaded bool
	listOut    string
	failVerbs  map[string]string
	installed  bool
}

func (s *darwinStub) install(t *testing.T) {
	t.Helper()
	oldRun := runDarwin
	oldInst := plistInstalled
	runDarwin = func(name string, args ...string) (string, error) {
		s.calls = append(s.calls, append([]string{name}, args...))
		if name != "launchctl" {
			return "", fmt.Errorf("unexpected command %q", name)
		}
		if len(args) == 0 {
			return "", fmt.Errorf("launchctl with no args")
		}
		verb := args[0]
		if msg, ok := s.failVerbs[verb]; ok {
			return "", fmt.Errorf("%s", msg)
		}
		if verb == "list" {
			if s.listLoaded {
				return s.listOut, nil
			}
			return "", fmt.Errorf("Could not find service")
		}
		return "", nil // load/unload succeed
	}
	plistInstalled = func(string) bool { return s.installed }
	t.Cleanup(func() {
		runDarwin = oldRun
		plistInstalled = oldInst
	})
}

func (s *darwinStub) call(idx int) []string {
	if len(s.calls) <= idx {
		return nil
	}
	return s.calls[idx]
}

func (s *darwinStub) last() []string {
	return s.call(len(s.calls) - 1)
}

// TestDarwinUsesLaunchctlWithPlistPath is the regression test for the macOS
// updater bug: the darwin build used to compile from svc_unix.go and call
// systemctl, which does not exist on macOS, so every self-update failed at
// "failed to stop service UTMStackAgent: exec: \"systemctl\": ... not found".
// The agent + updater are system launch daemons, so on darwin every control
// command must be launchctl, with load/unload addressed by the plist PATH and
// list by the bare label.
func TestDarwinUsesLaunchctlWithPlistPath(t *testing.T) {
	s := &darwinStub{installed: true}
	s.install(t)
	plist := daemonPlistPath("UTMStackAgent")

	if err := Stop("UTMStackAgent"); err != nil {
		t.Fatalf("Stop: unexpected error: %v", err)
	}
	if err := Start("UTMStackAgent"); err != nil {
		t.Fatalf("Start: unexpected error: %v", err)
	}
	if err := Restart("UTMStackAgent"); err != nil {
		t.Fatalf("Restart: unexpected error: %v", err)
	}

	// Stop(unload) + Start(load) + Restart(unload+load) = 4 commands.
	want := [][]string{
		{"launchctl", "unload", plist},
		{"launchctl", "load", plist},
		{"launchctl", "unload", plist},
		{"launchctl", "load", plist},
	}
	if !reflect.DeepEqual(s.calls, want) {
		t.Fatalf("commands =\n%v\nwant\n%v", s.calls, want)
	}

	// No command may reference systemctl (the original bug).
	for _, c := range s.calls {
		if c[0] == "systemctl" {
			t.Fatalf("found systemctl in darwin service calls: %v", c)
		}
	}
}

// TestDarwinIsActiveDetectsLoadedDaemon: launchctl list exits 0 with a PID in
// the output while the daemon is running.
func TestDarwinIsActiveDetectsLoadedDaemon(t *testing.T) {
	s := &darwinStub{installed: true, listLoaded: true, listOut: "PID\tStatus\tLastExitCode\n12345\t0\t\n"}
	s.install(t)

	if ok, err := IsActive("UTMStackAgent"); err != nil || !ok {
		t.Fatalf("IsActive: got (%v, %v), want (true, nil)", ok, err)
	}
	if c := s.last(); c[0] != "launchctl" || c[1] != "list" || c[2] != "UTMStackAgent" {
		t.Fatalf("IsActive used %v, want [launchctl list UTMStackAgent]", c)
	}
}

// TestDarwinIsActiveStoppedAndUnknown: a daemon that is loaded but idle shows
// no PID (stopped); a daemon that is not loaded errors from launchctl (still
// reported as not-running, without an error, since the updater health-checks
// right after Start).
func TestDarwinIsActiveStoppedAndUnknown(t *testing.T) {
	s := &darwinStub{installed: true, listLoaded: true, listOut: "PID\tStatus\tLastExitCode\n-\t0\t\n"}
	s.install(t)
	if ok, err := IsActive("UTMStackAgent"); err != nil || ok {
		t.Fatalf("IsActive (loaded, idle): got (%v, %v), want (false, nil)", ok, err)
	}

	s.listLoaded = false
	if ok, err := IsActive("UTMStackAgent"); err != nil || ok {
		t.Fatalf("IsActive (not loaded): got (%v, %v), want (false, nil)", ok, err)
	}
}

// TestDarwinStopOnUnloadedDaemonIsNoOp: launchctl unload fails for a daemon
// that is not currently loaded; that must not abort the update.
func TestDarwinStopOnUnloadedDaemonIsNoOp(t *testing.T) {
	s := &darwinStub{installed: true, listLoaded: false}
	s.failVerbs = map[string]string{"unload": "The service is not loaded"}
	s.install(t)

	if err := Stop("UTMStackAgent"); err != nil {
		t.Fatalf("Stop on unloaded daemon: %v", err)
	}
}

// TestDarwinStopFailsWhenDaemonStaysLoaded: if unload errors AND the daemon is
// still loaded, Stop must surface the error (it is not the benign no-op path).
func TestDarwinStopFailsWhenDaemonStaysLoaded(t *testing.T) {
	s := &darwinStub{installed: true, listLoaded: true}
	s.failVerbs = map[string]string{"unload": "unload failed"}
	s.install(t)

	if err := Stop("UTMStackAgent"); err == nil {
		t.Fatalf("Stop with daemon still loaded after unload error: want error, got nil")
	}
}

// TestDarwinMissingPlistIsAnError: a never-installed service has no plist;
// Start/Stop must fail rather than silently do nothing.
func TestDarwinMissingPlistIsAnError(t *testing.T) {
	s := &darwinStub{installed: false}
	s.install(t)

	if err := Start("NoSuchAgent"); err == nil {
		t.Fatalf("Start with missing plist: want error, got nil")
	}
	if err := Stop("NoSuchAgent"); err == nil {
		t.Fatalf("Stop with missing plist: want error, got nil")
	}
}

// TestDarwinStatusMapsLaunchdToServiceStates: running (PID present), stopped
// (loaded, no PID), stopped (not loaded but installed), unknown (not
// installed).
func TestDarwinStatusMapsLaunchdToServiceStates(t *testing.T) {
	s := &darwinStub{installed: true, listLoaded: true, listOut: "PID\tStatus\tLastExitCode\n12345\t0\t\n"}
	s.install(t)
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusRunning {
		t.Fatalf("Status running: got (%q, %v), want (running, nil)", st, err)
	}

	s.listOut = "PID\tStatus\tLastExitCode\n-\t0\t\n"
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusStopped {
		t.Fatalf("Status loaded-idle: got (%q, %v), want (stopped, nil)", st, err)
	}

	s.listLoaded = false
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusStopped {
		t.Fatalf("Status not-loaded: got (%q, %v), want (stopped, nil)", st, err)
	}

	s.installed = false
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusUnknown {
		t.Fatalf("Status not-installed: got (%q, %v), want (unknown, nil)", st, err)
	}
}

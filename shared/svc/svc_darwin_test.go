//go:build darwin

package svc

import (
	"fmt"
	"reflect"
	"testing"
)

// darwinStub is a fake command runner + plist-installed seam for the darwin
// service layer. It records every command and lets a test drive `launchctl
// list` (output + exit code) and force failures on load/unload.
type darwinStub struct {
	calls     [][]string
	listOut   string
	listExit  int
	installed bool
	failVerbs map[string]string
}

func (s *darwinStub) install(t *testing.T) {
	t.Helper()
	oldRun := runDarwin
	oldInst := plistInstalled
	runDarwin = func(name string, args ...string) (string, int, error) {
		s.calls = append(s.calls, append([]string{name}, args...))
		if name != "launchctl" || len(args) == 0 {
			return "", -1, fmt.Errorf("unexpected command %q %v", name, args)
		}
		verb := args[0]
		if msg, ok := s.failVerbs[verb]; ok {
			return "", 1, fmt.Errorf("%s", msg)
		}
		if verb == "list" {
			if s.listExit != 0 {
				return s.listOut, s.listExit, fmt.Errorf("exit %d", s.listExit)
			}
			return s.listOut, 0, nil
		}
		return "", 0, nil // load/unload succeed
	}
	plistInstalled = func(string) bool { return s.installed }
	t.Cleanup(func() {
		runDarwin = oldRun
		plistInstalled = oldInst
	})
}

func (s *darwinStub) last() []string {
	if len(s.calls) == 0 {
		return nil
	}
	return s.calls[len(s.calls)-1]
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

// TestDarwinStopSurfacesRealFailure is the HIGH-severity regression: if
// unload fails AND the daemon is still loaded/running (the follow-up list
// query shows a PID, or fails for a non-not-found reason), Stop must error -
// the updater must not swap the binary over a running agent.
func TestDarwinStopSurfacesRealFailure(t *testing.T) {
	// Case A: unload fails, list still shows a PID -> error.
	s := &darwinStub{installed: true, listOut: "PID\tStatus\tLastExitCode\n12345\t0\t\n"}
	s.failVerbs = map[string]string{"unload": "unload failed"}
	s.install(t)
	if err := Stop("UTMStackAgent"); err == nil {
		t.Fatalf("Stop (still running): want error, got nil")
	}

	// Case B: unload fails, list fails with a NON-not-found exit (e.g. 1)
	// -> error, because we cannot confirm the daemon is down.
	s2 := &darwinStub{installed: true, listExit: 1, listOut: "permission denied"}
	s2.failVerbs = map[string]string{"unload": "unload failed"}
	s2.install(t)
	if err := Stop("UTMStackAgent"); err == nil {
		t.Fatalf("Stop (list perm failure): want error, got nil")
	}
}

// TestDarwinStopOnUnloadedDaemonIsNoOp: launchctl unload fails for a daemon
// that is not currently loaded (list exits 113); that must not abort the
// update.
func TestDarwinStopOnUnloadedDaemonIsNoOp(t *testing.T) {
	s := &darwinStub{installed: true, listExit: 113}
	s.failVerbs = map[string]string{"unload": "The service is not loaded"}
	s.install(t)

	if err := Stop("UTMStackAgent"); err != nil {
		t.Fatalf("Stop on unloaded daemon: %v", err)
	}
}

// TestDarwinIsActiveDetectsLoadedDaemon: launchctl list exits 0 with a PID in
// the output while the daemon is running.
func TestDarwinIsActiveDetectsLoadedDaemon(t *testing.T) {
	s := &darwinStub{installed: true, listOut: "PID\tStatus\tLastExitCode\n12345\t0\t\n"}
	s.install(t)

	if ok, err := IsActive("UTMStackAgent"); err != nil || !ok {
		t.Fatalf("IsActive: got (%v, %v), want (true, nil)", ok, err)
	}
	if c := s.last(); c[0] != "launchctl" || c[1] != "list" || c[2] != "UTMStackAgent" {
		t.Fatalf("IsActive used %v, want [launchctl list UTMStackAgent]", c)
	}
}

// TestDarwinIsActiveStoppedAndNotFound: a loaded-but-idle daemon shows no PID;
// a not-loaded daemon (list exit 113) is not running without error.
func TestDarwinIsActiveStoppedAndNotFound(t *testing.T) {
	s := &darwinStub{installed: true, listOut: "PID\tStatus\tLastExitCode\n-\t0\t\n"}
	s.install(t)
	if ok, err := IsActive("UTMStackAgent"); err != nil || ok {
		t.Fatalf("IsActive (loaded, idle): got (%v, %v), want (false, nil)", ok, err)
	}

	s.listExit = 113
	s.listOut = ""
	if ok, err := IsActive("UTMStackAgent"); err != nil || ok {
		t.Fatalf("IsActive (not loaded): got (%v, %v), want (false, nil)", ok, err)
	}
}

// TestDarwinIsActiveSurfacesQueryFailure: a launchctl list failure that is not
// "not found" must propagate as an error, not be misreported as inactive.
func TestDarwinIsActiveSurfacesQueryFailure(t *testing.T) {
	s := &darwinStub{installed: true, listExit: 1, listOut: "permission denied"}
	s.install(t)

	if ok, err := IsActive("UTMStackAgent"); err == nil || ok {
		t.Fatalf("IsActive (perm failure): got (%v, %v), want (false, error)", ok, err)
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

// TestDarwinRejectsUnsafeServiceNames: names with path separators or
// traversal must be refused so the plist path can never escape
// /Library/LaunchDaemons.
func TestDarwinRejectsUnsafeServiceNames(t *testing.T) {
	s := &darwinStub{installed: true}
	s.install(t)

	for _, name := range []string{"../evil", "a/b", "a b", "", "-leading"} {
		if err := Start(name); err == nil {
			t.Fatalf("Start(%q): want error, got nil", name)
		}
		if err := Stop(name); err == nil {
			t.Fatalf("Stop(%q): want error, got nil", name)
		}
	}
}

// TestDarwinStatusMapsLaunchdToServiceStates: running (PID present), stopped
// (loaded, no PID), stopped (not loaded), unknown (not installed or query
// failure).
func TestDarwinStatusMapsLaunchdToServiceStates(t *testing.T) {
	s := &darwinStub{installed: true, listOut: "PID\tStatus\tLastExitCode\n12345\t0\t\n"}
	s.install(t)
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusRunning {
		t.Fatalf("Status running: got (%q, %v), want (running, nil)", st, err)
	}

	s.listOut = "PID\tStatus\tLastExitCode\n-\t0\t\n"
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusStopped {
		t.Fatalf("Status loaded-idle: got (%q, %v), want (stopped, nil)", st, err)
	}

	s.listExit = 113
	s.listOut = ""
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusStopped {
		t.Fatalf("Status not-loaded: got (%q, %v), want (stopped, nil)", st, err)
	}

	s.listExit = 1
	s.listOut = "permission denied"
	if st, err := Status("UTMStackAgent"); st != StatusUnknown || err == nil {
		t.Fatalf("Status query-failure: got (%q, %v), want (unknown, error)", st, err)
	}

	s.listExit = 0
	s.listOut = ""
	s.installed = false
	if st, err := Status("UTMStackAgent"); err != nil || st != StatusUnknown {
		t.Fatalf("Status not-installed: got (%q, %v), want (unknown, nil)", st, err)
	}
}

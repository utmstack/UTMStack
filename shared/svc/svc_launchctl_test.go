package svc

import "testing"

// plistRunningPID must recognize both launchctl `list` output shapes across
// macOS versions: the table form (PID in the first column) and the dictionary
// form used by older releases (a quoted "PID" key). It must return false for
// a registered-but-stopped daemon and for the "service not found" text
// launchctl prints for a daemon that is not loaded.
func TestPlistRunningPID(t *testing.T) {
	cases := []struct {
		name string
		out  string
		want bool
	}{
		{"table running", "PID\tStatus\tLastExitCode\n12345\t0\t\n", true},
		{"table running no header", "\n12345\t0\t\n", true},
		{"table stopped dash", "PID\tStatus\tLastExitCode\n-\t0\t\n", false},
		{"table stopped empty pid", "PID\tStatus\tLastExitCode\n\t0\t\n", false},
		{"dict running", "{\n\t\"PID\" = 12345;\n}", true},
		{"dict stopped", "{\n\t\"Program\" = \"/opt/utmstack-macos-agent/utmstack_agent_service\";\n\t\"LimitLoadToSessionType\" = \"Aqua\";\n}", false},
		{"not loaded message", "Could not find service \"UTMStackAgent\"", false},
		{"empty", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := plistRunningPID(tc.out); got != tc.want {
				t.Fatalf("plistRunningPID(%q) = %v, want %v", tc.out, got, tc.want)
			}
		})
	}
}

// launchctlListNotFound must classify `launchctl list <label>` results: the
// dedicated "not loaded" exit code (113) with no output, or any exit/output
// containing the human "could not find service" message (older launchctl
// versions print that text), is a not-found. A success (exit 0) or a
// permission/transient failure is NOT a not-found.
func TestLaunchctlListNotFound(t *testing.T) {
	cases := []struct {
		name      string
		exitCode  int
		out       string
		wantFound bool // want = NOT-found
	}{
		{"exit 113 no output", 113, "", true},
		{"exit 113 with output", 113, "Could not find service \"UTMStackAgent\"", true},
		{"exit 1 text notfound", 1, "Could not find service \"X\"", true},
		{"exit 0 loaded", 0, "PID\tStatus\n123\t0", false},
		{"exit 1 permission", 1, "Operation not permitted", false},
		{"exit 5 other failure", 5, "some other error", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := launchctlListNotFound(tc.exitCode, tc.out); got != tc.wantFound {
				t.Fatalf("launchctlListNotFound(%d, %q) = %v, want %v", tc.exitCode, tc.out, got, tc.wantFound)
			}
		})
	}
}

// daemonPlistPath is where kardianos/service writes a system launch daemon
// (optionUserService defaults to false): /Library/LaunchDaemons/<name>.plist.
// launchctl load/unload take that path, not the bare label.
func TestDaemonPlistPath(t *testing.T) {
	got := daemonPlistPath("UTMStackAgent")
	want := "/Library/LaunchDaemons/UTMStackAgent.plist"
	if got != want {
		t.Fatalf("daemonPlistPath(UTMStackAgent) = %q, want %q", got, want)
	}
}

// validLaunchdName accepts the identifier characters launchd labels and
// systemd unit names allow, and rejects anything with path separators or
// traversal so the plist path can never escape /Library/LaunchDaemons.
func TestValidLaunchdName(t *testing.T) {
	valid := []string{"UTMStackAgent", "UTMStackUpdater", "utmstack-collector", "svc.name-1_2", "A"}
	for _, name := range valid {
		if !validLaunchdName(name) {
			t.Fatalf("validLaunchdName(%q) = false, want true", name)
		}
	}
	invalid := []string{"", "UTMStackAgent/", "../evil", "a/b", "a\\b", "a b", "a:b", "-leading", ".hidden", "name*"}
	for _, name := range invalid {
		if validLaunchdName(name) {
			t.Fatalf("validLaunchdName(%q) = true, want false", name)
		}
	}
}

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

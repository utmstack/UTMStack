//go:build linux
// +build linux

package auditd

import (
	"fmt"
	"testing"
)

func execveEvent(argc int, args ...string) map[string]interface{} {
	execve := map[string]interface{}{"argc": itoa(argc)}
	for i, a := range args {
		execve[fmt.Sprintf("a%d", i)] = a
	}
	return map[string]interface{}{
		"syscall": "execve",
		"execve":  execve,
	}
}

func itoa(n int) string {
	return fmt.Sprintf("%d", n)
}

func TestIsNoiseEvent_RuncInitAndVersion(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"runc init", []string{"runc", "init"}},
		{"runc --version", []string{"runc", "--version"}},
		{"docker-init --version", []string{"/usr/libexec/docker/docker-init", "--version"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ev := execveEvent(len(c.args), c.args...)
			if !isNoiseEvent(ev) {
				t.Fatalf("expected %v to be noise", c.args)
			}
		})
	}
}

func TestIsNoiseEvent_KnownHealthchecks(t *testing.T) {
	cases := [][]string{
		{"curl", "-f", "http://localhost:8080/health"},
		{"curl", "-f", "http://localhost:8080/api/v1/health"},
		{"/bin/sh", "-c", "curl -f http://localhost:8080/health || exit 1"},
		{"/bin/sh", "-c", "curl -f http://localhost:8080/api/v1/health || exit 1"},
	}
	for _, args := range cases {
		ev := execveEvent(len(args), args...)
		if !isNoiseEvent(ev) {
			t.Fatalf("expected %v to be noise", args)
		}
	}
}

func TestIsNoiseEvent_DockerdRuncExecAndDelete(t *testing.T) {
	execArgs := []string{
		"runc", "--root", "/var/run/docker/runtime-runc/moby",
		"--log", "/run/containerd/.../log.json", "--log-format", "json",
		"--systemd-cgroup", "exec", "--process", "/tmp/runc-process123456",
	}
	if ev := execveEvent(len(execArgs), execArgs...); !isNoiseEvent(ev) {
		t.Fatalf("expected containerd-shim exec probe to be noise: %v", execArgs)
	}

	deleteArgs := []string{
		"runc", "--root", "/var/run/docker/runtime-runc/moby",
		"--log", "/run/containerd/.../log.json", "--log-format", "json",
		"--systemd-cgroup", "delete", "--force", "47dbe327b410",
	}
	if ev := execveEvent(len(deleteArgs), deleteArgs...); !isNoiseEvent(ev) {
		t.Fatalf("expected containerd-shim delete to be noise: %v", deleteArgs)
	}
}

func TestIsNoiseEvent_RuncPrivDropCascade(t *testing.T) {
	ev := map[string]interface{}{
		"syscall": "setuid",
		"comm":    "runc:[2:INIT]",
		"exe":     "/runc",
	}
	if !isNoiseEvent(ev) {
		t.Fatalf("expected runc's own init priv-drop syscall to be noise")
	}
}

func TestIsNoiseEvent_KnownPrivilegeDropBinaries(t *testing.T) {
	cases := []struct {
		name string
		comm string
		exe  string
	}{
		{"sudo", "sudo", "/usr/bin/sudo"},
		{"sshd", "sshd", "/usr/sbin/sshd"},
		{"systemd-executor", "systemd-executor", "/usr/lib/systemd/systemd-executor"},
		{"cron", "cron", "/usr/sbin/cron"},
		{
			// Real observed case: comm reflects a mid-transition
			// intermediate name (the sa1 script systemd-executor is about
			// to exec into) while exe still correctly reads
			// systemd-executor. Must still be excluded, matching on exe.
			"systemd-executor with misleading comm", "(sa1)", "/usr/lib/systemd/systemd-executor",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, syscall := range []string{"setuid", "setgid", "setresuid", "setresgid", "setreuid", "setregid"} {
				ev := map[string]interface{}{"syscall": syscall, "comm": tc.comm, "exe": tc.exe}
				if !isNoiseEvent(ev) {
					t.Fatalf("expected %s's own %s to be noise", tc.name, syscall)
				}
			}
		})
	}
}

func TestIsNoiseEvent_SudoOwnExecveDoesNotMatch(t *testing.T) {
	// The priv-drop exclusion must never reach execve -- the actual command
	// sudo ran is real signal and must still ship.
	args := []string{"sudo", "systemctl", "stop", "UTMStackAgent"}
	ev := execveEvent(len(args), args...)
	if isNoiseEvent(ev) {
		t.Fatalf("sudo's own execve record must not be filtered: %v", args)
	}
}

// --- Negative cases: things that must NOT be treated as noise ---

func TestIsNoiseEvent_UnknownBinaryPrivEscalationDoesNotMatch(t *testing.T) {
	// This is exactly the scenario the utmstack_priv audit rule exists to
	// catch: an unexpected binary (not one of the four known, verified
	// privilege-droppers) calling setuid. Must still ship.
	ev := map[string]interface{}{
		"syscall": "setuid",
		"comm":    "payload",
		"exe":     "/tmp/payload",
	}
	if isNoiseEvent(ev) {
		t.Fatalf("an unknown binary's setuid must not be filtered")
	}
}

func TestIsNoiseEvent_ManualRuncExecDoesNotMatch(t *testing.T) {
	// An attacker with host access to dockerd's own root dir, typing the
	// command by hand instead of going through containerd-shim's
	// --process/tmpfile invocation, must still ship.
	args := []string{
		"runc", "--root", "/var/run/docker/runtime-runc/moby",
		"exec", "47dbe327b410", "/bin/bash",
	}
	ev := execveEvent(len(args), args...)
	if isNoiseEvent(ev) {
		t.Fatalf("manual runc exec without --process tmpfile must not be filtered: %v", args)
	}
}

func TestIsNoiseEvent_UnknownCurlURLDoesNotMatch(t *testing.T) {
	args := []string{"curl", "-f", "http://169.254.169.254/latest/meta-data/"}
	ev := execveEvent(len(args), args...)
	if isNoiseEvent(ev) {
		t.Fatalf("curl to an unrecognized URL must not be filtered: %v", args)
	}
}

func TestIsNoiseEvent_CurlWithExtraArgsDoesNotMatch(t *testing.T) {
	// Same allow-listed URL, but with an extra flag (e.g. -o to save output)
	// -- shape no longer matches exactly, so it must still ship.
	args := []string{"curl", "-f", "-o", "/tmp/out", "http://localhost:8080/health"}
	ev := execveEvent(len(args), args...)
	if isNoiseEvent(ev) {
		t.Fatalf("curl with extra arguments must not be filtered: %v", args)
	}
}

func TestIsNoiseEvent_UnrelatedSetuidDoesNotMatch(t *testing.T) {
	// A comm of "sudo" alone, without the matching exe, must not be enough
	// -- exe is the field that's actually checked.
	ev := map[string]interface{}{
		"syscall": "setuid",
		"comm":    "sudo",
		"exe":     "/tmp/fake-sudo",
	}
	if isNoiseEvent(ev) {
		t.Fatalf("comm alone without a known exe must not be filtered")
	}
}

func TestIsNoiseEvent_RuncCommWithoutMatchingExeDoesNotMatch(t *testing.T) {
	// comm alone spoofed/coincidental without the corroborating exe path
	// must not be enough.
	ev := map[string]interface{}{
		"syscall": "setuid",
		"comm":    "runc:[2:INIT]",
		"exe":     "/usr/bin/some-other-binary",
	}
	if isNoiseEvent(ev) {
		t.Fatalf("comm without matching exe must not be filtered")
	}
}

func TestIsNoiseEvent_NonExecveNonPrivSyscallDoesNotMatch(t *testing.T) {
	ev := map[string]interface{}{"syscall": "openat"}
	if isNoiseEvent(ev) {
		t.Fatalf("unrelated syscalls must never be filtered")
	}
}

func TestIsNoiseEvent_MissingExecveMapDoesNotPanic(t *testing.T) {
	ev := map[string]interface{}{"syscall": "execve"}
	if isNoiseEvent(ev) {
		t.Fatalf("event without an execve object must not be filtered")
	}
}

func TestExecveArgs_MalformedArgcIsSafe(t *testing.T) {
	if got := execveArgs(map[string]interface{}{"argc": "not-a-number"}); got != nil {
		t.Fatalf("expected nil for malformed argc, got %v", got)
	}
	if got := execveArgs(map[string]interface{}{"argc": "0"}); got != nil {
		t.Fatalf("expected nil for zero argc, got %v", got)
	}
	if got := execveArgs(map[string]interface{}{"argc": "9999"}); got != nil {
		t.Fatalf("expected nil for implausible argc, got %v", got)
	}
}

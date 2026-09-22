//go:build linux
// +build linux

package auditd

import (
	"fmt"
	"strconv"
	"strings"
)

// isNoiseEvent: Docker container-lifecycle / healthcheck noise, verified
// against real ClickHouse data and every rules/linux/**/*.yaml where clause.
func isNoiseEvent(event map[string]interface{}) bool {
	syscall, _ := event["syscall"].(string)

	if isPrivSyscall(syscall) {
		return isKnownPrivDropSyscall(event)
	}

	if syscall != "execve" {
		return false
	}

	execve, ok := event["execve"].(map[string]interface{})
	if !ok {
		return false
	}
	return isNoiseExecve(execveArgs(execve))
}

func isPrivSyscall(syscall string) bool {
	switch syscall {
	case "setuid", "setgid", "setreuid", "setregid", "setresuid", "setresgid":
		return true
	}
	return false
}

// knownPrivilegeDropExes: binaries whose own priv-transition syscalls carry
// no signal beyond what's already captured elsewhere (sudo/cron -> their own
// execve record; sshd -> USER_START/CRED_ACQ; systemd-executor ->
// SERVICE_START/STOP). Matched on exe, not comm (comm can reflect a
// mid-transition intermediate name). utmstack_priv stays enabled for
// everything else, so an unknown binary calling setuid still ships.
var knownPrivilegeDropExes = map[string]bool{
	"/usr/bin/sudo":                     true,
	"/usr/sbin/sshd":                    true,
	"/usr/lib/systemd/systemd-executor": true,
	"/usr/sbin/cron":                    true,
}

// isKnownPrivDropSyscall covers runc's container-init privilege drop (comm
// is literally "runc:[2:INIT]", a value nothing else produces) plus
// knownPrivilegeDropExes.
func isKnownPrivDropSyscall(event map[string]interface{}) bool {
	exe, _ := event["exe"].(string)

	if exe == "/runc" || exe == "/usr/bin/runc" {
		comm, _ := event["comm"].(string)
		if comm == "runc:[2:INIT]" {
			return true
		}
	}

	return knownPrivilegeDropExes[exe]
}

// execveArgs extracts a0..a(argc-1) from an EXECVE record's fields, in
// order. Returns nil if argc is missing, non-numeric, or implausible.
func execveArgs(execve map[string]interface{}) []string {
	argcStr, _ := execve["argc"].(string)
	argc, err := strconv.Atoi(argcStr)
	if err != nil || argc <= 0 || argc > 128 {
		return nil
	}
	args := make([]string, argc)
	for i := 0; i < argc; i++ {
		args[i], _ = execve[fmt.Sprintf("a%d", i)].(string)
	}
	return args
}

// knownHealthcheckURLs: UTMStack's own container healthcheck endpoints
// (log-input, backend). Only matters when the agent watches a host that
// also runs the UTMStack stack itself.
var knownHealthcheckURLs = map[string]bool{
	"http://localhost:8080/health":        true,
	"http://localhost:8080/api/v1/health": true,
}

var knownHealthcheckShellCmds = map[string]bool{
	"curl -f http://localhost:8080/health || exit 1":        true,
	"curl -f http://localhost:8080/api/v1/health || exit 1": true,
}

// isNoiseExecve matches exact argv shapes: Docker/containerd's own
// container lifecycle calls (init, version probes, docker-exec healthcheck
// exec/delete) plus this host's two curl healthchecks. Extra or different
// args don't match and still ship.
func isNoiseExecve(args []string) bool {
	if len(args) == 0 {
		return false
	}

	switch args[0] {
	case "runc":
		if len(args) == 2 && (args[1] == "init" || args[1] == "--version") {
			return true
		}
		if len(args) >= 3 && args[1] == "--root" && strings.HasPrefix(args[2], "/var/run/docker/runtime-runc/") {
			return isDockerdRuncExecOrDelete(args[3:])
		}
	case "/usr/libexec/docker/docker-init":
		if len(args) == 2 && args[1] == "--version" {
			return true
		}
	case "curl":
		if len(args) == 3 && args[1] == "-f" && knownHealthcheckURLs[args[2]] {
			return true
		}
	case "/bin/sh":
		if len(args) == 3 && args[1] == "-c" && knownHealthcheckShellCmds[args[2]] {
			return true
		}
	}
	return false
}

// isDockerdRuncExecOrDelete recognizes containerd-shim's own invocation
// shape for `docker exec` (healthchecks) and post-exec teardown: a bare
// "delete", or "exec" paired with a "--process /tmp/runc-process*" flag --
// that flag is what containerd-shim itself adds, so a manually typed
// `runc exec <container> /bin/bash` won't match and still ships.
func isDockerdRuncExecOrDelete(rest []string) bool {
	for _, a := range rest {
		if a == "delete" {
			return true
		}
	}
	for i, a := range rest {
		if a != "exec" {
			continue
		}
		for j := i + 1; j < len(rest)-1; j++ {
			if rest[j] == "--process" && strings.HasPrefix(rest[j+1], "/tmp/runc-process") {
				return true
			}
		}
	}
	return false
}

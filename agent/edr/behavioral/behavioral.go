package behavioral

import (
	"path/filepath"
	"runtime"

	"github.com/utmstack/UTMStack/agent/edr/event"
	"github.com/utmstack/UTMStack/agent/edr/proctable"
)

// ProcTelemetry builds a behavioral telemetry event for a process start.
func ProcTelemetry(p proctable.Proc) event.Event {
	pc := event.ProcInfo{PID: p.PID, PPID: p.PPID, Image: p.Image, Cmdline: p.Cmdline}
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "process_create",
		ObjectPath: p.Image,
		Process:    &pc,
	}
}

// ScriptBlockTelemetry builds a behavioral event carrying deobfuscated script text.
func ScriptBlockTelemetry(scriptText, appName string) event.Event {
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "script_block",
		ObjectPath: appName,
		Signature:  truncate(scriptText, 8192),
	}
}

// ShellActivityTelemetry builds a behavioral event carrying what a shell or
// interpreter was asked to run: its command line, attributed to the
// interpreter name. This is the Linux counterpart of ScriptBlockTelemetry
// (the PowerShell 4104 record) and feeds the same correlation rules. It is
// platform-shared (a pure event builder); the IsInterpreter gate that decides
// WHEN it fires is linux-only.
func ShellActivityTelemetry(cmdline, interpreter string) event.Event {
	return event.Event{
		Source:     event.SourceBehavioral,
		Action:     "shell_activity",
		ObjectPath: interpreter,
		Signature:  truncate(cmdline, 8192),
	}
}

func truncate(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

// interpreterNames is the basename set of shells and interpreters whose
// execve we surface as shell_activity telemetry (the Linux counterpart of
// the PowerShell 4104 script-block records). A process whose image basename
// matches one of these emits shell_activity in addition to process_create.
//
// Deliberately conservative: only interactive/invoked interpreters. The
// shell builtins (echo, ls, ...) are child processes of these, so catching
// the interpreter's cmdline at execve covers what was "asked to run".
var interpreterNames = map[string]bool{
	// shells
	"bash": true, "sh": true, "dash": true, "zsh": true, "ksh": true,
	"ash": true, "busybox": true,
	// interpreted languages
	"python": true, "python2": true, "python3": true,
	"perl": true, "ruby": true, "php": true,
	"node": true, "deno": true, "bun": true,
	"lua": true, "tclsh": true, "wish": true,
	"expect": true, "pwsh": true, "powershell": true,
}

// IsInterpreter reports whether the image path refers to a known shell or
// interpreter, returning its basename alongside. It is Linux-only by design:
// the netlink process-connector watcher is the source of Linux execve, and
// the Windows 4104 script-block record already covers the equivalent there.
// A non-interpreter (or non-Linux) image returns ("", false) — the name is
// only meaningful on a match.
func IsInterpreter(image string) (string, bool) {
	if runtime.GOOS != "linux" {
		return "", false
	}
	return interpreterName(image)
}

// interpreterName is the platform-independent basename table. It is separated
// from IsInterpreter so the table can be unit-tested on any host; the
// linux-only gate lives in IsInterpreter.
func interpreterName(image string) (string, bool) {
	name := filepath.Base(image)
	if !interpreterNames[name] {
		return "", false
	}
	return name, true
}
